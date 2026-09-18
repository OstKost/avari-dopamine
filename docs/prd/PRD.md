---
doc_type: PRD
product_name: Dopamine Market
version: 1.0.0
status: approved
owner: product-owner
last_updated: 2026-09-18
target_scale:
  mau: 100
  concurrent_orders_peak: 10
  rps_peak: 20
machine_readable: true
schema_version: 1
---

# PRD: Dopamine Market

## 1. Резюме продукта

`product_id: dopamine-market`

Dopamine Market — веб-приложение, имитирующее опыт онлайн-маркетплейса и доставки еды, но специально спроектированное как **ритуал микро-удовлетворения**, а не как реальная торговая площадка. Пользователь регистрируется, выбирает пункт выдачи (случайный адрес в радиусе 100–500 м от геопозиции), ищет синтетические товары, наполняет корзину, оформляет заказ за фиксированную плату 10₽ (вне зависимости от состава корзины) и наблюдает симуляцию доставки/самовывоза через анимированный статус-трекер.

### 1.1 Проблема

```yaml
problem:
  id: P-01
  statement: >
    Людям иногда нужен быстрый, безопасный и дешёвый способ получить
    ощущение "покупки" и "предвкушения доставки" (дофаминовый цикл:
    поиск → выбор → ожидание → получение) без риска потратить крупную
    сумму денег и без реальных логистических издержек.
  evidence: hypothesis-driven, портфолио-проект без реального рынка на старте
```

### 1.2 Решение

```yaml
solution:
  id: S-01
  statement: >
    Маркетплейс-имитатор с фиксированной монетизацией 10 RUB за заказ,
    синтетическим бесконечным каталогом и симуляцией доставки через
    детерминированную стейт-машину по таймеру.
  non_goals:
    - Реальная логистика и курьеры
    - Реальные закупки/склад/остатки товаров
    - Комиссия, зависящая от суммы корзины
    - B2B/маркетплейс для продавцов (single-tenant каталог)
```

## 2. Целевая аудитория и масштаб

```yaml
scale_assumptions:
  initial_mau: 100
  avg_orders_per_user_per_day: 2
  avg_daily_orders: 200
  peak_rps: 20
  peak_concurrent_deliveries: 10
  data_retention_days: 90
  deployment: single-region, single Docker Compose host (portfolio-grade)
```

Проект физически рассчитан на 100 MAU: один Postgres-инстанс, один Kafka-брокер (KRaft, без Zookeeper), Redis без кластеризации. Архитектура должна быть горизонтально масштабируемой на бумаге (ADR фиксирует путь роста), но НЕ переусложнена инфраструктурно на старте — никакого k8s, никакого read-replica на 100 MAU.

## 3. Пользовательские роли

```yaml
actors:
  - id: guest
    description: Неавторизованный посетитель, видит каталог, не может заказывать
  - id: customer
    description: Зарегистрированный пользователь, основная роль
  - id: admin
    description: >
      Технический администратор (не отдельный UI на MVP, доступ через
      прямые SQL/скрипты и feature-флаги; полноценная admin-панель — backlog)
  - id: system_courier
    description: >
      Не человек. Виртуальная сущность, управляемая delivery state machine,
      не имеет собственного аккаунта
```

## 4. Функциональные требования (машиночитаемые)

Каждое требование — атомарная единица с `id`, используемым в Epics/Issues для трассируемости.

```yaml
functional_requirements:
  - id: FR-AUTH-01
    title: Регистрация по email + паролю
    priority: P0
    description: >
      Пользователь регистрируется email/паролем. Пароль хешируется argon2id.
      Подтверждение email — заглушка (лог вместо письма) на MVP.
    acceptance_criteria:
      - Email уникален (unique constraint, case-insensitive)
      - Пароль >= 8 символов, есть серверная валидация
      - После регистрации выдаётся сессия (не редиректит на верификацию)

  - id: FR-AUTH-02
    title: Вход и выход
    priority: P0
    description: Аутентификация по email+паролю, сессии на JWT access + refresh (httpOnly cookie).
    acceptance_criteria:
      - Access token TTL 15 минут, refresh TTL 30 дней
      - Refresh token — ротация при каждом использовании (reuse detection)
      - Logout инвалидирует refresh-сессию в Redis

  - id: FR-PICKUP-01
    title: Выбор пункта выдачи
    priority: P0
    description: >
      При онбординге пользователь даёт координаты (браузерная геолокация
      либо ручной ввод города). Система генерирует 5-8 синтетических
      пунктов выдачи со случайным смещением 100-500м от заданной точки
      (случайный угол + расстояние, проекция через геодезическую формулу).
    acceptance_criteria:
      - Смещение всегда в диапазоне [100, 500] метров
      - Названия ПВЗ генерируются из шаблонов (см. DATA-GEN-02)
      - Пользователь может сменить активный ПВЗ в любой момент до оформления заказа
      - Если геолокация недоступна — fallback на координаты города по умолчанию (Rostov-on-Don)

  - id: FR-CATALOG-01
    title: Каталог товаров с категориями
    priority: P0
    description: Синтетический каталог 150-300 товаров, 8-12 категорий, генерируется сидом при старте.
    acceptance_criteria:
      - Товар имеет name, category, price_rub (для UI реализма, не участвует в оплате), image_seed, description
      - Каталог статичен между релизами (детерминированный сид для воспроизводимости)

  - id: FR-CATALOG-02
    title: Поиск и фильтрация товаров
    priority: P1
    description: Full-text поиск по названию/описанию (Postgres tsvector), фильтр по категории.
    acceptance_criteria:
      - Поиск возвращает результат < 200мс на 300 товарах
      - Пустой запрос возвращает дефолтную выдачу по категориям

  - id: FR-CART-01
    title: Корзина пользователя
    priority: P0
    description: >
      Корзина хранится в Redis (ключ по user_id), TTL 7 дней неактивности.
      Один активный ПВЗ на корзину.
    acceptance_criteria:
      - Добавление/удаление/изменение количества товара
      - Корзина видна между сессиями (persisted в Redis, не localStorage)
      - При смене ПВЗ корзина не сбрасывается (ПВЗ не привязан к товарам)

  - id: FR-ORDER-01
    title: Оформление заказа
    priority: P0
    description: >
      Из корзины создаётся заказ. Оплата фиксированная 10 RUB через
      PaymentProvider интерфейс (см. ADR-006). Состав корзины копируется
      в order_items как снапшот (цены каталога могут меняться, заказ хранит
      цену на момент покупки).
    acceptance_criteria:
      - Пустая корзина не может быть оформлена
      - После создания заказа корзина очищается
      - Заказ создаётся в статусе `created`, платёж инициируется асинхронно
      - Публикуется событие order.created (outbox)

  - id: FR-ORDER-02
    title: Жизненный цикл статусов заказа
    priority: P0
    description: >
      Статусная машина: created → payment_pending → paid → assembling →
      courier_assigned → in_transit → delivered. Ветка отказа:
      payment_failed, cancelled.
    acceptance_criteria:
      - Невалидные переходы статуса отклоняются на уровне домена (не только API)
      - Все переходы логируются с timestamp в order_status_history
      - Пользователь получает обновления в реальном времени (SSE)

  - id: FR-PAY-01
    title: Единый интерфейс оплаты
    priority: P0
    description: >
      Домен payment работает через порт PaymentProvider с двумя адаптерами:
      MockProvider (детерминированная имитация с 95% success rate) и
      YooKassaProvider (реальная интеграция, тестовый режим/sandbox).
      Выбор адаптера — через конфигурацию окружения, без изменений домена.
    acceptance_criteria:
      - Сумма платежа всегда 10.00 RUB, валюта RUB
      - MockProvider эмулирует webhook callback с задержкой 1-3 сек
      - YooKassaProvider обрабатывает реальные webhook с проверкой подписи
      - Оба адаптера реализуют идентичный Go-интерфейс `payment.Provider`

  - id: FR-DELIVERY-01
    title: Симуляция доставки/самовывоза
    priority: P0
    description: >
      После успешной оплаты запускается стейт-машина курьера, управляемая
      таймером (см. ADR-005). Каждый переход — отдельное событие Kafka,
      транслируемое клиенту через SSE. Общая длительность доставки —
      случайная величина 90-240 секунд.
    acceptance_criteria:
      - Этапы: assembling (10-30с) → courier_assigned → in_transit (60-180с) → delivered
      - Курьер имеет синтетическое имя и "рейтинг" (генерируется при назначении)
      - Прогресс отображается в UI в процентах и с ETA

  - id: FR-NOTIF-01
    title: Real-time обновления статуса заказа
    priority: P0
    description: SSE-эндпоинт транслирует изменения статуса конкретного заказа клиенту.
    acceptance_criteria:
      - Переподключение клиента восстанавливает текущий статус (не только будущие события)
      - Задержка доставки события клиенту < 500мс от публикации в Kafka

  - id: FR-HISTORY-01
    title: История заказов
    priority: P1
    description: Список прошлых заказов пользователя с фильтром по статусу и датой.
    acceptance_criteria:
      - Пагинация курсорная (created_at, id)
      - Каждый заказ можно "повторить" (создать новую корзину с тем же составом)

  - id: FR-GAMIFY-01
    title: Дофаминовые UI-элементы
    priority: P1
    description: >
      Анимация прогресса доставки, confetti/микро-анимация при получении
      заказа, счётчик "всего заказов" пользователя, streak (дни подряд с заказом).
    acceptance_criteria:
      - Streak считается по календарным дням в TZ пользователя
      - Анимации не блокируют доступность (respects prefers-reduced-motion)
```

## 5. Нефункциональные требования

```yaml
non_functional_requirements:
  - id: NFR-PERF-01
    category: performance
    requirement: "p95 latency API < 200ms при 20 RPS"
  - id: NFR-OBS-01
    category: observability
    requirement: >
      Все сервисы пишут структурные JSON-логи (slog) с trace_id,
      экспортируются в Graylog через GELF UDP/TCP input
  - id: NFR-OBS-02
    category: observability
    requirement: "Каждый HTTP-запрос и Kafka-consumer имеют distributed trace (OpenTelemetry)"
  - id: NFR-REL-01
    category: reliability
    requirement: >
      Публикация событий через outbox pattern — потеря событий недопустима
      при рестарте сервиса
  - id: NFR-REL-02
    category: reliability
    requirement: "Kafka consumers идемпотентны (дедупликация по event_id)"
  - id: NFR-SEC-01
    category: security
    requirement: "Пароли — argon2id, JWT подписаны, webhook YooKassa проверяет HMAC подпись"
  - id: NFR-SEC-02
    category: security
    requirement: "Rate limiting на auth endpoints (5 попыток/мин на IP) через Redis"
  - id: NFR-MAINT-01
    category: maintainability
    requirement: >
      Модули backend взаимодействуют только через явные Go-интерфейсы
      в internal/contracts — запрещён прямой импорт internal-пакетов
      другого модуля (проверяется линтером go-arch-lint / depguard)
  - id: NFR-COST-01
    category: cost
    requirement: "Вся инфраструктура работает на одной VPS (4 vCPU/8GB) через Docker Compose"
```

## 6. Ключевые бизнес-правила (инварианты)

```yaml
business_invariants:
  - id: INV-01
    rule: "Плата за заказ ВСЕГДА 10.00 RUB, независимо от количества/стоимости товаров в корзине"
  - id: INV-02
    rule: "Один пользователь не может иметь два заказа в статусе payment_pending одновременно"
  - id: INV-03
    rule: "Пункт выдачи всегда находится в радиусе 100-500м от координат, заданных пользователем"
  - id: INV-04
    rule: "Заказ неизменяем после перехода в paid (нельзя менять состав)"
  - id: INV-05
    rule: "Каждое доменное событие публикуется ровно один раз потребителю (exactly-once semantics через outbox + idempotency key)"
```

## 7. Метрики успеха

```yaml
success_metrics:
  - id: METRIC-01
    name: orders_per_active_user_per_week
    target: ">= 5"
  - id: METRIC-02
    name: day_7_retention
    target: ">= 30%"
  - id: METRIC-03
    name: order_completion_rate
    description: "created -> delivered без cancelled/payment_failed"
    target: ">= 90%"
  - id: METRIC-04
    name: p95_checkout_latency
    target: "< 300ms"
```

## 8. Явные допущения и границы (для агентов)

```yaml
assumptions:
  - "Каталог и курьеры — 100% синтетические данные, генерируются сидами при первом запуске (см. Epic CATALOG-SEED)"
  - "Геопозиция ПВЗ — фейковые координаты со смещением, не привязаны к реальным картам/адресному API"
  - "Симуляция доставки — таймер-driven state machine, БЕЗ реального геотрекинга/маршрутов на карте (см. ADR-005)"
  - "На 100 MAU допустима single-instance инфраструктура без HA — это осознанное упрощение, задокументированное в ADR-002"
  - "Реальный платёжный провайдер (YooKassa) подключается в тестовом/sandbox режиме — реальные деньги не проходят"
```

## 9. Трассируемость к Epics

Каждый `FR-*`/`NFR-*` должен быть покрыт хотя бы одной задачей в `docs/epics/`. Полная матрица трассируемости — `docs/epics/traceability.md` (генерируется/обновляется агентом при декомпозиции).
