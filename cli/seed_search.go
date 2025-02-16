package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/TommasoAmici/fountain-finder/pkg/osm"
	"github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
)

var rdb *redis.Client
var rdbCache *cache.Cache

// This program should be run to seed the cache for the geocoder.
// It will fetch the results for all permutations of 3 characters from `aaa` to `zzz`
// and store them in redis for a month.
// Since there is a cooldown of 1s between each request it takes a total of 5 hours to run.
// https://operations.osmfoundation.org/policies/nominatim/

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

	userAgent := os.Getenv("USER_AGENT")

	for ch1 := 'a'; ch1 < 'z'; ch1++ {
		search(fmt.Sprintf("%c", ch1), userAgent)
		for ch2 := 'a'; ch2 < 'z'; ch2++ {
			search(fmt.Sprintf("%c%c", ch1, ch2), userAgent)
			for ch3 := 'a'; ch3 < 'z'; ch3++ {
				search(fmt.Sprintf("%c%c%c", ch1, ch2, ch3), userAgent)
			}
		}
	}
}

func search(query string, userAgent string) {
	fmt.Println("Fetching ", query)
	cacheKey := fmt.Sprintf("osm:%s", query)
	var result []osm.GeoCodeResponse
	err := rdbCache.Get(context.Background(), cacheKey, &result)
	if err == nil {
		return
	}

	time.Sleep(1 * time.Second)

	result, err = osm.Geocode(query, userAgent)
	if err != nil {
		return
	}

	rdbCache.Set(&cache.Item{
		Ctx:   context.Background(),
		Key:   cacheKey,
		Value: result,
		TTL:   60 * 24 * time.Hour,
	})
}
