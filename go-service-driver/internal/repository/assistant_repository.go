package repository

import (
	"context"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

type assistantRepository struct {
	rdb *redis.Client
}

func (ar *assistantRepository) GetSummary(ctx context.Context, userID uint64) (string, error) {
	return ar.rdb.Get(ctx, ar.key(userID)).Result()
}

func (ar *assistantRepository) SetSummary(ctx context.Context, userID uint64, summary string) error {
	return ar.rdb.Set(ctx, ar.key(userID), summary, 24*time.Hour).Err()
}

func NewAssistantRepository(rdb *redis.Client) AssistantRepository {
	return &assistantRepository{rdb: rdb}
}

func (ar *assistantRepository) key(userID uint64) string {
	return "assistant:summary:" + strconv.FormatUint(userID, 10)
}
