package repository

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

const assistantHistoryTTL = 24 * time.Hour

type assistantRepository struct {
	rdb *redis.Client
}

func NewAssistantRepository(rdb *redis.Client) AssistantRepository {
	return &assistantRepository{rdb: rdb}
}

func (ar *assistantRepository) GetSummary(ctx context.Context, userID uint64) (string, error) {
	return ar.rdb.Get(ctx, ar.summaryKey(userID)).Result()
}

func (ar *assistantRepository) SetSummary(ctx context.Context, userID uint64, summary string) error {
	return ar.rdb.Set(ctx, ar.summaryKey(userID), summary, 24*time.Hour).Err()
}

func (ar *assistantRepository) AppendTurn(ctx context.Context, userID uint64, turn AssistantTurn, limit int) error {
	payload, err := json.Marshal(turn)
	if err != nil {
		return err
	}
	key := ar.historyKey(userID)
	pipe := ar.rdb.Pipeline()
	pipe.LPush(ctx, key, payload)
	if limit > 0 {
		pipe.LTrim(ctx, key, 0, int64(limit-1))
	}
	pipe.Expire(ctx, key, assistantHistoryTTL)
	_, err = pipe.Exec(ctx)
	return err
}

func (ar *assistantRepository) GetRecentTurns(ctx context.Context, userID uint64, limit int) ([]AssistantTurn, error) {
	if limit <= 0 {
		return nil, nil
	}
	raws, err := ar.rdb.LRange(ctx, ar.historyKey(userID), 0, int64(limit-1)).Result()
	if err != nil {
		return nil, err
	}
	// LPUSH 写入,LRANGE 0..N-1 是新→旧,反转回时间正序
	turns := make([]AssistantTurn, 0, len(raws))
	for i := len(raws) - 1; i >= 0; i-- {
		var t AssistantTurn
		if err := json.Unmarshal([]byte(raws[i]), &t); err != nil {
			continue
		}
		turns = append(turns, t)
	}
	return turns, nil
}

func (ar *assistantRepository) AcquireSummaryLock(ctx context.Context, userID uint64, token string, ttl time.Duration) (bool, error) {
	return ar.rdb.SetNX(ctx, ar.lockKey(userID), token, ttl).Result()
}

// releaseLockScript 仅在 value 与 token 匹配时删除,防止误删别人重新获取的锁
var releaseLockScript = redis.NewScript(`
if redis.call("get", KEYS[1]) == ARGV[1] then
	return redis.call("del", KEYS[1])
else
	return 0
end
`)

func (ar *assistantRepository) ReleaseSummaryLock(ctx context.Context, userID uint64, token string) error {
	return releaseLockScript.Run(ctx, ar.rdb, []string{ar.lockKey(userID)}, token).Err()
}

func (ar *assistantRepository) summaryKey(userID uint64) string {
	return "assistant:summary:" + strconv.FormatUint(userID, 10)
}

func (ar *assistantRepository) historyKey(userID uint64) string {
	return "assistant:history:" + strconv.FormatUint(userID, 10)
}

func (ar *assistantRepository) lockKey(userID uint64) string {
	return "assistant:summarize:lock:" + strconv.FormatUint(userID, 10)
}
