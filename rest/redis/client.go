package redis_client

import "github.com/redis/go-redis/v9"

var (
	Client redis.Client
)

func InitRedisClient() {
	Client = *redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Replace with your Redis server address
		Password: "",               // No password set
		DB:       0,                // Use default DB
		Protocol: 2,
	})
}

func GetRedisClient() *redis.Client {
	if Client.Options().Addr == "" {
		InitRedisClient()
	}
	return &Client
}
