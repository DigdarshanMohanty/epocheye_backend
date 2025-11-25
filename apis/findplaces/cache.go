package findplaces

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()
var Rdb *redis.Client

func InitRedis() {
	Rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	if _, err := Rdb.Ping(ctx).Result(); err != nil {
		panic("❌ Redis not reachable: " + err.Error())
	}
	fmt.Println("✅ Connected to Redis")
}

// func CacheResponse(key string, data FindPlacesResponse) error {
// 	bytes, err := json.Marshal(data)
// 	if err != nil {
// 		return err
// 	}
// 	return Rdb.Set(ctx, key, bytes, 5*time.Minute).Err()
// }

//	func GetCachedResponse(key string) (*FindPlacesResponse, bool) {
//		val, err := Rdb.Get(ctx, key).Result()
//		if err != nil {
//			return nil, false
//		}
//		var res FindPlacesResponse
//		if err := json.Unmarshal([]byte(val), &res); err != nil {
//			return nil, false
//		}
//		return &res, true
//	}
func SetCachedResponse(key string, value interface{}, duration time.Duration) {
	bytes, _ := json.Marshal(value)
	Rdb.Set(ctx, key, bytes, duration)
}
