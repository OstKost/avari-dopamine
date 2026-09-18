# Epics — декомпозиция для агентной разработки

Каждый Epic — самодостаточная единица работы для одного AI-агента (или последовательной серии запусков), с явными зависимостями, областью файлов (`scope.paths`) и acceptance criteria, трассируемыми к `docs/prd/PRD.md`.

## Граф зависимостей

```
EPIC-00 (Platform Bootstrap)
   │
   ├──> EPIC-01 (Identity & Auth)
   │        │
   ├──> EPIC-02 (Catalog & Synthetic Data)
   │        │
   ├──> EPIC-03 (Pickup Points)
   │        │
   │        ▼
   │   EPIC-04 (Cart)
   │        │
   │        ▼
   │   EPIC-05 (Order Lifecycle) <─────┐
   │        │                          │
   │        ▼                          │
   │   EPIC-06 (Payment Abstraction) ──┘
   │        │
   │        ▼
   │   EPIC-07 (Delivery Simulation)
   │        │
   │        ▼
   │   EPIC-08 (Realtime Notifications / SSE)
   │
   ├──> EPIC-09 (Frontend Shell & Design System)
   │        │ (зависит от EPIC-01, 02, 03 API-контрактов)
   │        ▼
   │   EPIC-10 (Frontend Catalog & Cart UX)
   │        │
   │        ▼
   │   EPIC-11 (Frontend Checkout & Order Tracking UX)
   │
   ├──> EPIC-12 (Observability: Logging, Tracing, Metrics)
   ├──> EPIC-13 (CI/CD & Quality Gates)
   └──> EPIC-14 (Gamification Layer)
```

`EPIC-00` — обязательная предпосылка для всех остальных. `EPIC-01/02/03` можно вести параллельно тремя агентами сразу после `EPIC-00`. `EPIC-09` (frontend shell) можно начинать параллельно с backend-эпиками после того, как `openapi.yaml` содержит хотя бы skeleton-контракты.

## Реестр Epics

| ID | Название | Модуль(и) | Приоритет | Зависит от |
|---|---|---|---|---|
| [EPIC-00](EPIC-00-platform-bootstrap.md) | Platform Bootstrap | platform, infra | P0 | — |
| [EPIC-01](EPIC-01-identity-auth.md) | Identity & Auth | identity | P0 | EPIC-00 |
| [EPIC-02](EPIC-02-catalog.md) | Catalog & Synthetic Data | catalog | P0 | EPIC-00 |
| [EPIC-03](EPIC-03-pickup-points.md) | Pickup Points | pickup | P0 | EPIC-00 |
| [EPIC-04](EPIC-04-cart.md) | Cart | cart | P0 | EPIC-01, EPIC-02, EPIC-03 |
| [EPIC-05](EPIC-05-order-lifecycle.md) | Order Lifecycle | order | P0 | EPIC-04 |
| [EPIC-06](EPIC-06-payment.md) | Payment Abstraction | payment | P0 | EPIC-05 |
| [EPIC-07](EPIC-07-delivery.md) | Delivery Simulation | delivery | P0 | EPIC-06 |
| [EPIC-08](EPIC-08-notifications.md) | Realtime Notifications (SSE) | notification | P0 | EPIC-05, EPIC-07 |
| [EPIC-09](EPIC-09-frontend-shell.md) | Frontend Shell & Design System | web | P0 | EPIC-00, (contracts из 01/02/03) |
| [EPIC-10](EPIC-10-frontend-catalog-cart.md) | Frontend Catalog & Cart UX | web | P0 | EPIC-09, EPIC-02, EPIC-03, EPIC-04 |
| [EPIC-11](EPIC-11-frontend-checkout-tracking.md) | Frontend Checkout & Order Tracking | web | P0 | EPIC-09, EPIC-05, EPIC-08 |
| [EPIC-12](EPIC-12-observability.md) | Observability | platform | P1 | EPIC-00 |
| [EPIC-13](EPIC-13-ci-cd.md) | CI/CD & Quality Gates | infra | P1 | EPIC-00 |
| [EPIC-14](EPIC-14-gamification.md) | Gamification Layer | order, web | P2 | EPIC-05, EPIC-11 |

## Матрица трассируемости к PRD

| PRD ID | Epic(s) |
|---|---|
| FR-AUTH-01, FR-AUTH-02 | EPIC-01 |
| FR-PICKUP-01 | EPIC-03 |
| FR-CATALOG-01, FR-CATALOG-02 | EPIC-02 |
| FR-CART-01 | EPIC-04 |
| FR-ORDER-01, FR-ORDER-02 | EPIC-05 |
| FR-PAY-01 | EPIC-06 |
| FR-DELIVERY-01 | EPIC-07 |
| FR-NOTIF-01 | EPIC-08 |
| FR-HISTORY-01 | EPIC-05, EPIC-11 |
| FR-GAMIFY-01 | EPIC-14 |
| NFR-PERF-01 | EPIC-13 (нагрузочный тест в CI) |
| NFR-OBS-01, NFR-OBS-02 | EPIC-12 |
| NFR-REL-01, NFR-REL-02 | EPIC-05, EPIC-06, EPIC-07 (outbox/idempotency) |
| NFR-SEC-01, NFR-SEC-02 | EPIC-01 |
| NFR-MAINT-01 | EPIC-00 (lint-arch), все Epics соблюдают |
| NFR-COST-01 | EPIC-00 (docker-compose) |

## Как агенту брать Epic в работу

1. Прочитать Epic-файл целиком, включая `scope.paths` и `acceptance_criteria`.
2. Проверить, что все `depends_on` Epics имеют статус `done` (смотреть `status` в front-matter — обновляется вручную/через PR merge).
3. Создать ветку `feat/{epic-id}-{slug}`.
4. Работать строго в `scope.paths`. Если нужно выйти за границы — см. `AGENTS.md` раздел 8.
5. Выполнить все `tasks` из Epic по порядку (это рекомендованная, не строго обязательная последовательность внутри Epic).
6. Пройти Definition of Done из `AGENTS.md` раздела 5 + специфичный для Epic `definition_of_done`.
