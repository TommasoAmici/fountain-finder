package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/logger"
	"github.com/kataras/iris/v12/middleware/recover"
)

var rdb *redis.Client
var rdbCache *cache.Cache

//go:embed dist/*
var embedWeb embed.FS

func main() {
	redisAddress, found := os.LookupEnv("ADDRESS")
	if !found {
		redisAddress = "127.0.0.1:6379"
	}
	rdb = redis.NewClient(&redis.Options{
		Addr:     redisAddress,
		Password: "",
		DB:       0,
	})

	rdbCache = cache.New(&cache.Options{
		Redis: rdb,
	})

	app := newApp()
	addr, found := os.LookupEnv("ADDRESS")
	if !found {
		addr = "127.0.0.1:8000"
	}
	app.Listen(addr)
}

func newApp() *iris.Application {
	app := iris.New()

	app.UseGlobal(logger.New(), recover.New())

	app.Get("/api/health", healthCheck)
	app.Get("/api/{startLng}/{startLat}/{endLng}/{endLat}", getHandler)

	fsys := iris.PrefixDir("dist", http.FS(embedWeb))
	app.RegisterView(iris.HTML(fsys, ".html"))
	app.HandleDir("/", fsys)

	return app
}

func healthCheck(ctx iris.Context) {
	ctx.StatusCode(iris.StatusOK)
}

type Element struct {
	ID  int     `json:"id" msgpack:"id"`
	Lat float64 `json:"lat" msgpack:"lat"`
	Lng float64 `json:"lon" msgpack:"lon"`
}

type Response struct {
	Elements []Element `json:"elements" msgpack:"elements"`
}

func fetchElements(startLng, startLat, endLng, endLat float64) (map[string]interface{}, error) {
	baseURL := "https://overpass.kumi.systems/api/interpreter/?data="
	bounds := fmt.Sprintf("(%f,%f,%f,%f)", startLat, startLng, endLat, endLng)
	query := fmt.Sprintf(`[out:json][timeout:25];(node["amenity"="drinking_water"]%s;);out body;>;out skel qt;`, bounds)
	u := baseURL + url.QueryEscape(query)
	r, err := http.Get(u)
	if err != nil {
		return nil, err
	}

	data := Response{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&data)
	if err != nil {
		return nil, err
	}
	r.Body.Close()

	result := map[string]interface{}{}
	result["elements"] = data.Elements
	return result, nil
}

func getHandler(ctx iris.Context) {
	startLng := ctx.Params().GetFloat64Default("startLng", 0)
	startLat := ctx.Params().GetFloat64Default("startLat", 0)
	endLng := ctx.Params().GetFloat64Default("endLng", 0)
	endLat := ctx.Params().GetFloat64Default("endLat", 0)

	cacheLng := math.Trunc(startLng)
	cacheLat := math.Trunc(startLat)
	cacheKey := fmt.Sprintf("overpass:%f-%f", cacheLng, cacheLat)
	var result map[string]interface{}
	rdbCache.Get(ctx.Request().Context(), cacheKey, &result)
	if result != nil {
		ctx.Header("Cache-Control", "public, max-age=3600, stale-while-revalidate=86400")
		ctx.JSON(result)
		return
	}
	go func() {
		result, err := fetchElements(cacheLng, cacheLat, cacheLng+1, cacheLat+1)
		if err != nil {
			ctx.Application().Logger().Error("Failed to fetch JSON from Overpass API: ", cacheKey)
			return
		}
		rdbCache.Set(&cache.Item{
			Ctx:   ctx.Request().Context(),
			Key:   cacheKey,
			Value: result,
			TTL:   24 * time.Hour,
		})
	}()

	result, err := fetchElements(startLng, startLat, endLng, endLat)
	if err != nil {
		ctx.Application().Logger().Error("Failed to get elements from Overpass API")
		prob := iris.NewProblem().Detail("Failed to get elements from Overpass API").Status(iris.StatusInternalServerError)
		ctx.Problem(prob)
		return
	}

	ctx.Header("Cache-Control", "public, max-age=3600, stale-while-revalidate=86400")
	ctx.JSON(result)
}
