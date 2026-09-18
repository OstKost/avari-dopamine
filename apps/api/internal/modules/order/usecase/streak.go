package usecase

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/ostkost/dopamine-market/api/internal/modules/order/domain"
)

type UserStats struct {
	TotalOrders       int        `json:"total_orders"`
	CurrentStreakDays int        `json:"current_streak_days"`
	LongestStreakDays int        `json:"longest_streak_days"`
	TotalSavedRUB     string     `json:"total_saved_rub"`
	LastOrderAt       *time.Time `json:"last_order_at,omitempty"`
}

// GetUserStats рассчитывает статистику заказов и streak пользователя по дням в его TZ (FR-GAMIFY-01).
func (uc *OrderUseCase) GetUserStats(ctx context.Context, userID uuid.UUID, tzName string) (UserStats, error) {
	loc := time.UTC
	if tzName != "" {
		if l, err := time.LoadLocation(tzName); err == nil {
			loc = l
		}
	}

	orders, _, err := uc.repo.ListByUserID(ctx, userID, 1000, 0)
	if err != nil {
		return UserStats{}, fmt.Errorf("listing orders for stats: %w", err)
	}

	if len(orders) == 0 {
		return UserStats{
			TotalOrders:       0,
			CurrentStreakDays: 0,
			LongestStreakDays: 0,
			TotalSavedRUB:     "0.00",
		}, nil
	}

	// Фильтруем успешные/оплаченные заказы
	var validOrders []*domain.Order
	for _, o := range orders {
		if o.Status() != domain.StatusCancelled && o.Status() != domain.StatusPaymentFailed {
			validOrders = append(validOrders, o)
		}
	}

	if len(validOrders) == 0 {
		return UserStats{
			TotalOrders:       len(orders),
			CurrentStreakDays: 0,
			LongestStreakDays: 0,
			TotalSavedRUB:     "0.00",
		}, nil
	}

	// Сортируем по возрастанию даты создания
	sort.Slice(validOrders, func(i, j int) bool {
		return validOrders[i].CreatedAt().Before(validOrders[j].CreatedAt())
	})

	lastOrderTime := validOrders[len(validOrders)-1].CreatedAt()

	// Извлекаем уникальные календарные дни в TZ пользователя
	datesMap := make(map[string]time.Time)
	for _, o := range validOrders {
		localTime := o.CreatedAt().In(loc)
		dayKey := localTime.Format("2006-01-02")
		datesMap[dayKey] = time.Date(localTime.Year(), localTime.Month(), localTime.Day(), 0, 0, 0, 0, loc)
	}

	var uniqueDays []time.Time
	for _, d := range datesMap {
		uniqueDays = append(uniqueDays, d)
	}
	sort.Slice(uniqueDays, func(i, j int) bool {
		return uniqueDays[i].Before(uniqueDays[j])
	})

	// Подсчет streak
	currentStreak := 0
	longestStreak := 0

	if len(uniqueDays) > 0 {
		tempStreak := 1
		longestStreak = 1

		for i := 1; i < len(uniqueDays); i++ {
			prev := uniqueDays[i-1]
			curr := uniqueDays[i]

			// Разница ровно в 1 календарный день
			if curr.Sub(prev).Hours() <= 26 && curr.Day() != prev.Day() {
				tempStreak++
				if tempStreak > longestStreak {
					longestStreak = tempStreak
				}
			} else {
				tempStreak = 1
			}
		}

		// Проверяем актуальность current streak относительно сегодня/вчера
		nowLocal := time.Now().In(loc)
		today := time.Date(nowLocal.Year(), nowLocal.Month(), nowLocal.Day(), 0, 0, 0, 0, loc)
		yesterday := today.AddDate(0, 0, -1)

		lastActiveDay := uniqueDays[len(uniqueDays)-1]
		if lastActiveDay.Equal(today) || lastActiveDay.Equal(yesterday) {
			currentStreak = tempStreak
		} else {
			currentStreak = 0
		}
	}

	return UserStats{
		TotalOrders:       len(validOrders),
		CurrentStreakDays: currentStreak,
		LongestStreakDays: longestStreak,
		TotalSavedRUB:     fmt.Sprintf("%d.00", len(validOrders)*190), // Иллюстративная синтетическая экономия
		LastOrderAt:       &lastOrderTime,
	}, nil
}
