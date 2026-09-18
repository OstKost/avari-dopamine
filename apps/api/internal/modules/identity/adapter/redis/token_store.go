package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ostkost/dopamine-market/api/internal/modules/identity/domain"
	"github.com/redis/go-redis/v9"
)

type TokenStore struct {
	client *redis.Client
}

func NewTokenStore(client *redis.Client) *TokenStore {
	return &TokenStore{client: client}
}

func (s *TokenStore) SaveToken(ctx context.Context, familyID, tokenID, userID uuid.UUID, ttl time.Duration) error {
	key := fmt.Sprintf("auth:family:%s", familyID.String())
	pipe := s.client.Pipeline()
	pipe.HSet(ctx, key, map[string]interface{}{
		"active_token_id": tokenID.String(),
		"user_id":         userID.String(),
		"revoked":         "false",
	})
	pipe.Expire(ctx, key, ttl)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("saving refresh token family to redis: %w", err)
	}
	return nil
}

// luaRotateScript реализует атомарную проверку и ротацию токена с защитой от race condition и reuse-detection.
const luaRotateScript = `
local key = KEYS[1]
local presented_token_id = ARGV[1]
local new_token_id = ARGV[2]
local user_id = ARGV[3]
local ttl_seconds = tonumber(ARGV[4])

local exists = redis.call('EXISTS', key)
if exists == 0 then
    return -1 -- not found / expired
end

local revoked = redis.call('HGET', key, 'revoked')
if revoked == 'true' then
    return -2 -- already revoked
end

local active_token_id = redis.call('HGET', key, 'active_token_id')
if active_token_id ~= presented_token_id then
    -- Reuse detected! Инвалидируем всю family
    redis.call('HSET', key, 'revoked', 'true')
    redis.call('EXPIRE', key, 300) -- держим 5 минут для повторных детектов
    return -3 -- reuse detected
end

-- Успешная ротация
redis.call('HSET', key, 'active_token_id', new_token_id)
redis.call('EXPIRE', key, ttl_seconds)
return 1
`

func (s *TokenStore) ValidateAndRotate(ctx context.Context, familyID, tokenID, newID, userID uuid.UUID, ttl time.Duration) error {
	key := fmt.Sprintf("auth:family:%s", familyID.String())
	res, err := s.client.Eval(ctx, luaRotateScript, []string{key},
		tokenID.String(),
		newID.String(),
		userID.String(),
		int(ttl.Seconds()),
	).Int()

	if err != nil {
		return fmt.Errorf("evaluating token rotation script: %w", err)
	}

	switch res {
	case 1:
		return nil
	case -1:
		return domain.ErrInvalidRefreshToken
	case -2:
		return domain.ErrInvalidRefreshToken
	case -3:
		return domain.ErrRefreshTokenReused
	default:
		return domain.ErrInvalidRefreshToken
	}
}

func (s *TokenStore) RevokeFamily(ctx context.Context, familyID uuid.UUID) error {
	key := fmt.Sprintf("auth:family:%s", familyID.String())
	err := s.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("revoking refresh token family: %w", err)
	}
	return nil
}
