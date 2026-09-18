package  cache

import(
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
)

var Client *redis.Client

func Connect(addr string, password string, db int)error{
	Client = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx := context.Background()
	if err:= Client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}
	fmt.Println("Connected to Redis")
	return nil

}