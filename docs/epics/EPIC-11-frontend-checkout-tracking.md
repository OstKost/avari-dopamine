---
id: EPIC-11
title: Frontend Checkout & Order Tracking
module: [web]
priority: P0
status: todo
depends_on: [EPIC-09, EPIC-05, EPIC-08]
covers_requirements: [FR-ORDER-01, FR-ORDER-02, FR-NOTIF-01, FR-HISTORY-01]
related_adrs: [ADR-009]
scope:
  paths:
    - apps/web/app/(main)/checkout/**
    - apps/web/app/(main)/orders/**
    - apps/web/components/features/order/**
    - apps/web/hooks/useOrderStatus.ts
---

# EPIC-11: Frontend Checkout & Order Tracking

## Цель

Оформление заказа и главный "дофаминовый" экран — real-time трекинг статуса заказа с анимированным прогрессом, плюс история заказов. Это витрина всей event-driven архитектуры для конечного пользователя.

## Задачи

1. `app/(main)/checkout/page.tsx`: подтверждение состава корзины + ПВЗ, кнопка "Оформить за 10₽" → Server Action, вызывающий `POST /orders`, редирект на `/orders/{id}`.
2. `app/(main)/orders/[id]/page.tsx`: Server Component — SSR первого рендера статуса (без "мигания" пустого состояния, ADR-009), передаёт начальные данные клиентскому компоненту-обёртке.
3. `hooks/useOrderStatus.ts`: клиентский hook, оборачивающий `EventSource` на `/api/orders/{id}/events`, обрабатывает события `snapshot`/`status_changed`, автопереподключение браузера "из коробки" + логика обработки `onerror`.
4. `components/features/order/DeliveryProgress.tsx`: визуальный прогресс-бар/таймлайн стадий (`assembling → courier_assigned → in_transit → delivered`), с плавной анимацией перехода между стадиями (CSS transitions/Framer Motion), отображение ETA, имени и рейтинга курьера при `courier_assigned`.
5. `components/features/order/DeliveryDelayedBanner.tsx`: отдельное состояние для `delivery_delayed` ветки — лёгкое, не раздражающее уведомление ("курьер немного задержался") — сохраняет вовлечённость, не создаёт негатива.
6. Момент `delivered`: микро-анимация (confetti/celebration, respects `prefers-reduced-motion` — NFR из PRD FR-GAMIFY-01, реализуется базово здесь, расширяется в EPIC-14).
7. `app/(main)/orders/page.tsx`: история заказов, Server Component с курсорной пагинацией (`?cursor=`), карточка заказа с датой/статусом/кнопкой "Повторить".
8. Обработка сетевых сбоев SSE: если `EventSource` не может установить соединение (например корпоративный proxy блокирует) — fallback на polling каждые 5с (deграциозная деградация, не критичный баг, но хорошая демонстрация продуманности).

## Acceptance Criteria

- [ ] Оформление заказа переводит на страницу трекинга без ручного обновления (`FR-ORDER-01`)
- [ ] Статус на странице заказа обновляется в реальном времени без перезагрузки страницы (`FR-NOTIF-01`)
- [ ] При обновлении страницы (F5) статус восстанавливается корректно (не сбрасывается на "assembling") (`FR-NOTIF-01`)
- [ ] История заказов пагинируется, "Повторить" наполняет корзину прежним составом (`FR-HISTORY-01`)
- [ ] Анимации уважают `prefers-reduced-motion`

## Definition of Done

Стандартный DoD. Записать короткий screen-recording (GIF) полного цикла оформление→трекинг→доставлено для README/презентации портфолио — это самый демонстрационный экран продукта.
