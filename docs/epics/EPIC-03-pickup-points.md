---
id: EPIC-03
title: Pickup Points
module: [pickup]
priority: P0
status: done
depends_on: [EPIC-00]
covers_requirements: [FR-PICKUP-01, INV-03]
related_adrs: [ADR-012]
scope:
  paths:
    - apps/api/internal/modules/pickup/**
    - apps/api/internal/contracts/pickup.go
    - apps/api/migrations/pickup/**
    - docs/api/openapi.yaml   # секция /pickup/*
---

# EPIC-03: Pickup Points

## Цель

По координатам пользователя сгенерировать 5-8 синтетических пунктов выдачи в радиусе 100-500м (ADR-012, INV-03) и позволить пользователю выбрать активный ПВЗ.

## Задачи

1. `domain/pickup_point.go`: сущность `PickupPoint` (id, user_id, name, latitude, longitude, distance_meters, created_at). ПВЗ генерируются per-user (не общий каталог точек), что проще для изоляции данных и соответствует "любой адрес в 100-500м от пользователя".
2. `domain/geo.go`: чистая функция `DestinationPoint(origin LatLng, bearingDeg float64, distanceMeters float64) LatLng` — геодезическая формула на сфере (ADR-012). Юнит-тесты с известными эталонными значениями (например расстояние от точки на экваторе).
3. `port/`: `PickupPointRepository`, `RandomSource` (из `internal/platform/random`, инжектируется, не импортируется напрямую — соблюдение AGENTS.md раздел 7).
4. `usecase/GeneratePickupPoints`: принимает координаты пользователя, генерирует 5-8 точек с `bearing ~ Uniform(0,360)`, `distance ~ Uniform(100,500)`, валидирует результат строго в диапазоне (защитный тест на границы).
5. `usecase/SelectPickupPoint`: устанавливает активный ПВЗ пользователя (хранится как FK на `pickup_points.id` в контексте пользователя — где именно хранить "активный" статус целесообразно решить как часть корзины, см. ADR-008: `cart:{user_id}:pickup_point` в Redis, значение — `pickup_point_id`; этот Epic предоставляет только генерацию и чтение точек, запись "активного" делает EPIC-04).
6. `adapter/postgres`: схема `pickup`, таблица `pickup_points`.
7. `adapter/httpapi`: `POST /pickup/generate` (body: {latitude, longitude} ИЛИ {city: "Rostov-on-Don"} для fallback без геолокации — `FR-PICKUP-01` fallback требование), `GET /pickup/points` (список уже сгенерированных для пользователя).
8. Название точек — шаблонизированная генерация (`"Пункт выдачи №{N}"`, `"Шкафчик у {ориентир}"`) из пула шаблонов, ориентиры — заготовленный список нейтральных слов (не претендующих на реальные городские объекты, чтобы не создавать путаницу с реальными адресами).
9. Тесты: unit на `DestinationPoint` (граничные случаи, полюса — не обязательно, но 0/90/180/270 градусов), unit на диапазон distance/bearing генерации (property-based: 1000 прогонов, все в допустимом диапазоне), интеграционный тест API.

## Acceptance Criteria

- [x] Сгенерированные точки всегда в диапазоне [100, 500] метров от заданных координат (`INV-03`) — property-based тест обязателен
- [x] Генерируется 5-8 точек за один вызов (`FR-PICKUP-01`)
- [x] При недоступности геолокации — fallback на координаты Ростова-на-Дону (`FR-PICKUP-01`)
- [x] Пользователь может сменить активный ПВЗ до оформления заказа (обеспечивается вместе с EPIC-04)

## Definition of Done

Стандартный DoD. Особое внимание: `DestinationPoint` — чистая, детально протестированная функция, так как это единственная нетривиальная математика в проекте — хороший кандидат для демонстрации качества unit-тестирования в портфолио.
