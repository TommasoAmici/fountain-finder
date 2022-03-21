package main

import (
	"context"
	"flag"
	"fmt"
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
// Because there is a cooldown of 1s between each request it takes a total of 5 hours to run.
// https://operations.osmfoundation.org/policies/nominatim/

var redisAddress *string
var userAgent *string

func init() {
	redisAddress = flag.String("redis", "127.0.0.1:6379", "redis address")
	userAgent = flag.String("ua", "", "user agent for API requests")
}

func main() {
	flag.Parse()

	rdb = redis.NewClient(&redis.Options{
		Addr:     *redisAddress,
		Password: "",
		DB:       0,
	})

	rdbCache = cache.New(&cache.Options{
		Redis: rdb,
	})

	for ch1 := 'a'; ch1 < 'z'; ch1++ {
		search(fmt.Sprintf("%c", ch1))
		for ch2 := 'a'; ch2 < 'z'; ch2++ {
			search(fmt.Sprintf("%c%c", ch1, ch2))
			for ch3 := 'a'; ch3 < 'z'; ch3++ {
				search(fmt.Sprintf("%c%c%c", ch1, ch2, ch3))
			}
		}
	}
}

func search(query string) {
	fmt.Println("Fetching ", query)
	cacheKey := fmt.Sprintf("osm:%s", query)
	var result []osm.GeoCodeResponse
	err := rdbCache.Get(context.Background(), cacheKey, &result)
	if err == nil {
		return
	}

	time.Sleep(1 * time.Second)

	result, err = osm.Geocode(query, *userAgent)
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
