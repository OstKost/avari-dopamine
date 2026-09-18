---
id: EPIC-04
title: Cart
module: [cart]
priority: P0
status: todo
depends_on: [EPIC-01, EPIC-02, EPIC-03]
covers_requirements: [FR-CART-01]
related_adrs: [ADR-008]
scope:
  paths:
    - apps/api/internal/modules/cart/**
    - apps/api/internal/contracts/cart.go
    - docs/api/openapi.yaml   # секция /cart/*
---

# EPIC-04: Cart

## Цель

Redis-backed корзина (ADR-008): добавление/удаление/изменение количества товаров, привязка активного ПВЗ, TTL 7 дней с продлением.

## Задачи

1. `domain/cart.go`: value object `Cart` (items map[ProductID]Quantity, pickup_point_id nullable) — иммутабельный, методы `WithItem`, `WithoutItem`, `WithPickupPoint` возвращают новую копию (функциональный стиль упрощает тестирование).
2. `port/CartStore`: интерфейс Get/Save/Clear/Touch(продление TTL), реализуется в `adapter/redis`.
3. `usecase/`: `AddItem`, `RemoveItem`, `UpdateQuantity`, `SetPickupPoint`, `GetCart` (обогащает товарами из `contracts.ProductLookup` — актуальные названия/цены на момент просмотра корзины, в отличие от заказа, где это снапшот).
4. При `AddItem`/`SetPickupPoint` — валидация существования `product_id`/`pickup_point_id` через соответствующие contracts-интерфейсы (`catalog.ProductLookup`, `pickup.PickupPointLookup`) — ошибка, если товар/точка не существует.
5. `adapter/redis`: hash-структура `cart:{user_id}`, отдельный ключ `cart:{user_id}:pickup`, `EXPIRE` при каждой мутации (TTL 7 дней).
6. `adapter/httpapi`: `GET /cart`, `POST /cart/items`, `PATCH /cart/items/{product_id}`, `DELETE /cart/items/{product_id}`, `PUT /cart/pickup-point`. Все под `RequireAuth` middleware (EPIC-01).
7. Тесты: unit на domain (`Cart` иммутабельность, edge cases — добавление нулевого количества, удаление несуществующего товара), usecase с mock-портами, интеграционный тест TTL-продления (проверка, что `EXPIRE` действительно обновляется).

## Acceptance Criteria

- [ ] Добавление/удаление/изменение количества работает корректно (`FR-CART-01`)
- [ ] Корзина сохраняется между сессиями/устройствами одного пользователя (server-side, не localStorage) (`FR-CART-01`)
- [ ] Смена ПВЗ не сбрасывает товары в корзине (`FR-CART-01`)
- [ ] Добавление несуществующего `product_id` возвращает явную ошибку `404`, а не тихо игнорируется

## Definition of Done

Стандартный DoD. Дополнительно: нагрузочный smoke-тест (100 конкурентных мутаций одной корзины) — не строгий SLA, но демонстрация отсутствия race condition (Redis-операции атомарны по дизайну — подтвердить, что реализация не вводит несогласованность через read-modify-write без транзакции/Lua-скрипта, где это важно).
