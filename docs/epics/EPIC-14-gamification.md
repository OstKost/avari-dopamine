---
id: EPIC-14
title: Gamification Layer
module: [order, web]
priority: P2
status: todo
depends_on: [EPIC-05, EPIC-11]
covers_requirements: [FR-GAMIFY-01]
related_adrs: []
scope:
  paths:
    - apps/api/internal/modules/order/usecase/streak.go
    - apps/api/internal/modules/order/adapter/httpapi/stats.go
    - apps/web/components/features/gamification/**
    - apps/web/app/(main)/profile/**
---

# EPIC-14: Gamification Layer

## Цель

Дополнительные "дофаминовые" элементы: счётчик всего заказов, streak (дни подряд с заказом), более выразительные микро-анимации. Это P2 — реализуется после основного цикла заказа полностью работает end-to-end.

## Задачи

1. Backend: `usecase/GetUserStats` в модуле `order` — считает `total_orders`, `current_streak_days` (по календарным дням в TZ пользователя — требует хранения TZ пользователя или использования TZ из заголовка запроса/профиля, зафиксировать подход при реализации: проще брать `Intl.DateTimeFormat().resolvedOptions().timeZone` на клиенте и передавать как query param, чем хранить TZ в БД для портфолио-масштаба).
2. `adapter/httpapi`: `GET /users/me/stats`.
3. Frontend: `components/features/gamification/StreakBadge.tsx` (отображается в шапке/профиле), `components/features/gamification/OrderCounter.tsx`.
4. `app/(main)/profile/page.tsx`: простая страница профиля со статистикой (если не создана ранее в других Epics — проверить перед началом, не дублировать).
5. Усилить celebration-анимацию из EPIC-11 (`delivered` момент) — добавить контекст ("Это твой 12-й заказ!" при round numbers, streak-уведомление "3 дня подряд!").
6. Все анимации — проверить `prefers-reduced-motion` (повторное явное требование, так как это P2-Epic легко забывает про accessibility при добавлении "весёлых" фич).

## Acceptance Criteria

- [ ] Streak считается корректно по календарным дням, сбрасывается при пропуске дня (`FR-GAMIFY-01`)
- [ ] Счётчик заказов отображается и обновляется после каждого нового заказа
- [ ] Анимации не блокируют доступность (`prefers-reduced-motion` соблюдается)

## Definition of Done

Стандартный DoD. Так как это P2 — допустимо смержить с минимальным визуальным полишем и итерировать позже, но backend-логика streak должна быть покрыта unit-тестами на граничные случаи (пропуск дня, разные часовые зоны, заказ ровно в полночь).
