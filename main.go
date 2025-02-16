package main

import (
	"embed"
	"fmt"
	"math"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/TommasoAmici/fountain-finder/pkg/osm"
	"github.com/getsentry/sentry-go"
	sentryiris "github.com/getsentry/sentry-go/iris"
	"github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/logger"
	"github.com/kataras/iris/v12/middleware/recover"
)

var rdb *redis.Client
var rdbCache *cache.Cache

var GitCommit string

//go:embed dist/*
var embedWeb embed.FS

func main() {
	redisAddress, found := os.LookupEnv("REDIS_ADDRESS")
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

	if err := sentry.Init(sentry.ClientOptions{
		Dsn:              os.Getenv("SENTRY_DSN"),
		EnableTracing:    true,
		TracesSampleRate: 1.0,
	}); err != nil {
		fmt.Printf("Sentry initialization failed: %v\n", err)
	}

	app := newApp()
	app.Use(sentryiris.New(sentryiris.Options{}))
	app.Listen("0.0.0.0:8000")
}

func newApp() *iris.Application {
	app := iris.New()

	app.UseGlobal(logger.New(), recover.New())

	app.Get("/api/health", healthCheck)
	app.Get("/api/search", searchHandler)
	app.Get("/api/fountains/{startLng}/{startLat}/{endLng}/{endLat}", getHandler)

	env := os.Getenv("ENV")
	if env == "PRODUCTION" {
		fsys := iris.PrefixDir("dist", http.FS(embedWeb))
		app.RegisterView(iris.HTML(fsys, ".html"))
		app.HandleDir("/", fsys)
	}

	return app
}

func healthCheck(ctx iris.Context) {
	ctx.StatusCode(iris.StatusOK)
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
		result, err := osm.FetchElements(cacheLng, cacheLat, cacheLng+1, cacheLat+1)
		if err != nil {
			ctx.Application().Logger().Error("Failed to fetch JSON from Overpass API: ", cacheKey)
			return
		}
		result["cache"] = true
		rdbCache.Set(&cache.Item{
			Ctx:   ctx.Request().Context(),
			Key:   cacheKey,
			Value: result,
			TTL:   24 * time.Hour,
		})
	}()

	result, err := osm.FetchElements(startLng, startLat, endLng, endLat)
	if err != nil {
		ctx.Application().Logger().Error("Failed to get elements from Overpass API")
		prob := iris.NewProblem().Detail("Failed to get elements from Overpass API").Status(iris.StatusInternalServerError)
		ctx.Problem(prob)
		return
	}

	ctx.Header("Cache-Control", "public, max-age=3600, stale-while-revalidate=86400")
	ctx.JSON(result)
}

var cooldown = false

func searchHandler(ctx iris.Context) {
	query := ctx.URLParam("query")
	if query == "" {
		prob := iris.NewProblem().Detail("Empty queries are not supported").Status(iris.StatusBadRequest)
		ctx.Problem(prob)
		return
	}
	maxLength := 15
	if len(query) > maxLength {
		query = query[:maxLength]
	}
	query = strings.ToLower(query)

	cacheKey := fmt.Sprintf("osm:%s", query)
	var result []osm.GeoCodeResponse
	err := rdbCache.Get(ctx.Request().Context(), cacheKey, &result)
	if err == nil {
		ctx.Header("Cache-Control", "public, max-age=86400, stale-while-revalidate=86400")
		ctx.JSON(result)
		return
	}

	// wait 1s between API requests
	// https://operations.osmfoundation.org/policies/nominatim/
	if cooldown {
		time.Sleep(1 * time.Second)
		cooldown = false
	}
	cooldown = true

	result, err = osm.Geocode(query, "fountain-finder")
	if err != nil {
		ctx.Application().Logger().Error("Failed to geocode ", query, err)
		prob := iris.NewProblem().Detail("Failed to geocode query").Status(iris.StatusInternalServerError)
		ctx.Problem(prob)
		return
	}

	rdbCache.Set(&cache.Item{
		Ctx:   ctx.Request().Context(),
		Key:   cacheKey,
		Value: result,
		TTL:   24 * time.Hour,
	})

	ctx.Header("Cache-Control", "public, max-age=86400, stale-while-revalidate=86400")
	ctx.JSON(result)
}
