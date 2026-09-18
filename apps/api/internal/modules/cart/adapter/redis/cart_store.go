package redis

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/ostkost/dopamine-market/api/internal/modules/cart/domain"
	"github.com/redis/go-redis/v9"
)

type CartStore struct {
	client *redis.Client
}

func NewCartStore(client *redis.Client) *CartStore {
	return &CartStore{client: client}
}

func (s *CartStore) GetCart(ctx context.Context, userID uuid.UUID) (domain.Cart, error) {
	itemsKey := fmt.Sprintf("cart:%s:items", userID.String())
	pickupKey := fmt.Sprintf("cart:%s:pickup", userID.String())

	pipe := s.client.Pipeline()
	itemsCmd := pipe.HGetAll(ctx, itemsKey)
	pickupCmd := pipe.Get(ctx, pickupKey)
	_, err := pipe.Exec(ctx)
	if err != nil && !errors.Is(err, redis.Nil) {
		return domain.Cart{}, fmt.Errorf("getting cart from redis: %w", err)
	}

	itemsMap := make(map[uuid.UUID]int)
	rawItems, _ := itemsCmd.Result()
	for k, v := range rawItems {
		prodID, err := uuid.Parse(k)
		if err != nil {
			continue
		}
		qty, err := strconv.Atoi(v)
		if err != nil || qty <= 0 {
			continue
		}
		itemsMap[prodID] = qty
	}

	var pickupID *uuid.UUID
	rawPickup, err := pickupCmd.Result()
	if err == nil && rawPickup != "" {
		if id, err := uuid.Parse(rawPickup); err == nil {
			pickupID = &id
		}
	}

	return domain.NewCart(itemsMap, pickupID), nil
}

func (s *CartStore) SaveCart(ctx context.Context, userID uuid.UUID, cart domain.Cart, ttl time.Duration) error {
	itemsKey := fmt.Sprintf("cart:%s:items", userID.String())
	pickupKey := fmt.Sprintf("cart:%s:pickup", userID.String())

	pipe := s.client.Pipeline()

	// 1. Items
	pipe.Del(ctx, itemsKey)
	items := cart.Items()
	if len(items) > 0 {
		hashData := make(map[string]interface{}, len(items))
		for k, v := range items {
			hashData[k.String()] = v
		}
		pipe.HSet(ctx, itemsKey, hashData)
		pipe.Expire(ctx, itemsKey, ttl)
	}

	// 2. Pickup Point
	if cart.PickupPointID() != nil {
		pipe.Set(ctx, pickupKey, cart.PickupPointID().String(), ttl)
	} else {
		pipe.Del(ctx, pickupKey)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("saving cart to redis: %w", err)
	}

	return nil
}

func (s *CartStore) ClearCart(ctx context.Context, userID uuid.UUID) error {
	itemsKey := fmt.Sprintf("cart:%s:items", userID.String())
	pickupKey := fmt.Sprintf("cart:%s:pickup", userID.String())

	pipe := s.client.Pipeline()
	pipe.Del(ctx, itemsKey)
	pipe.Del(ctx, pickupKey)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("clearing cart in redis: %w", err)
	}
	return nil
}

func (s *CartStore) Touch(ctx context.Context, userID uuid.UUID, ttl time.Duration) error {
	itemsKey := fmt.Sprintf("cart:%s:items", userID.String())
	pickupKey := fmt.Sprintf("cart:%s:pickup", userID.String())

	pipe := s.client.Pipeline()
	pipe.Expire(ctx, itemsKey, ttl)
	pipe.Expire(ctx, pickupKey, ttl)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("touching cart ttl in redis: %w", err)
	}
	return nil
}
