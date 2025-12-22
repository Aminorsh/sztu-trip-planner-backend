package repository

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type assistantRepository struct {
	rdb *redis.Client
}

func (ar *assistantRepository) GetSummary(ctx context.Context, userID string) (string, error) {
	return ar.rdb.Get(ctx, ar.key(userID)).Result()
}

func (ar *assistantRepository) SetSummary(ctx context.Context, userID string, summary string) error {
	return ar.rdb.Set(ctx, ar.key(userID), summary, 24*time.Hour).Err()
}

func NewAssistantRepository(rdb *redis.Client) AssistantRepository {
	return &assistantRepository{rdb: rdb}
}

func (ar *assistantRepository) key(userID string) string {
	return "assistant:summary:" + userID
}
