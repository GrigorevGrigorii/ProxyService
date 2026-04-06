package cache

import (
	"context"
	"errors"
	"fmt"
	"proxy-service/internal/models"
	"strconv"

	"github.com/redis/go-redis/v9"
)

type Repository interface {
	Set(ctx context.Context, service models.ServiceDTO, target models.TargetDTO, data string, statusCode int, contentType string) error
	Get(ctx context.Context, service models.ServiceDTO, target models.TargetDTO) (string, int, string, error)
}

type CacheRepository struct {
	Redis *redis.UniversalClient
}

func (r *CacheRepository) Set(ctx context.Context, service models.ServiceDTO, target models.TargetDTO, data string, statusCode int, contentType string) error {
	return (*r.Redis).HSet(
		ctx, cacheKey(service, target),
		map[string]any{
			"data":         data,
			"status_code":  statusCode,
			"content_type": contentType,
		},
	).Err()
}

func (r *CacheRepository) Get(ctx context.Context, service models.ServiceDTO, target models.TargetDTO) (string, int, string, error) {
	redisResult, err := (*r.Redis).HGetAll(ctx, cacheKey(service, target)).Result()
	if err != nil {
		return "", 0, "", err
	}
	if len(redisResult) == 0 {
		return "", 0, "", errors.New("Data not found")
	}

	data := redisResult["data"]
	contentType := redisResult["content_type"]
	statusCode, err := strconv.Atoi(redisResult["status_code"])
	if err != nil {
		return "", 0, "", err
	}

	return data, statusCode, contentType, nil
}

func cacheKey(service models.ServiceDTO, target models.TargetDTO) string {
	return fmt.Sprintf("%s:%s:%s:%s", service.Name, target.Path, target.Method, target.Query)
}
