# ADR-006: Абстракция PaymentProvider (Mock + YooKassa)

- Статус: accepted
- Дата: 2026-09-18
- Контекст PRD: FR-PAY-01, INV-01. Продуктовое решение пользователя: "оба варианта через единый интерфейс".

## Контекст

Пользователь явно запросил и мок-провайдер (для быстрой демонстрации без внешних зависимостей и ключей API), и интеграцию с реальным платёжным провайдером в тестовом режиме (для демонстрации умения работать с внешним API, webhook, подписями). Оба должны быть переключаемы без изменения домена `order`/`payment`.

## Решение

### Порт

```go
// internal/contracts/payment.go
package contracts

type PaymentProvider interface {
    // Initiate начинает платёж, возвращает провайдер-специфичный идентификатор
    // и (опционально) URL для редиректа пользователя (для YooKassa).
    Initiate(ctx context.Context, req InitiatePaymentRequest) (InitiatePaymentResult, error)

    // VerifyWebhook проверяет подпись/подлинность входящего вебхука и
    // возвращает нормализованный результат (не зависящий от провайдера).
    VerifyWebhook(ctx context.Context, raw []byte, headers http.Header) (WebhookResult, error)
}

type InitiatePaymentRequest struct {
    OrderID     uuid.UUID
    AmountRUB   decimal.Decimal // всегда 10.00, валидируется на уровне usecase (INV-01)
    IdempotencyKey string
}
```

### Адаптеры

1. **MockProvider** (`adapter/mock`): `Initiate` сразу возвращает `status=pending`, планирует само-вызов webhook через тот же delivery-scheduler механизм (ADR-005) с задержкой 1-3с и success rate 95% (5% — детерминированный `payment_failed` для демонстрации error-path в UI). Не требует внешней сети — используется по умолчанию в `docker-compose` и CI.
2. **YooKassaProvider** (`adapter/yookassa`): реальный HTTP-клиент к YooKassa API (`https://api.yookassa.ru/v3/payments`) в тестовом магазине (shop_id из sandbox), `VerifyWebhook` проверяет IP-подпись согласно документации YooKassa. Требует секретов (`YOOKASSA_SHOP_ID`, `YOOKASSA_SECRET_KEY`) — если не заданы, приложение падает при старте с явной ошибкой конфигурации (fail-fast), а не тихим фоллбэком на mock.

### Выбор адаптера

Через `PAYMENT_PROVIDER=mock|yookassa` в конфиге (`internal/platform/config`). Composition root (`cmd/server/main.go`) собирает нужный адаптер и инжектит как `contracts.PaymentProvider` в модуль `order`.

### Идемпотентность

`Initiate` принимает `IdempotencyKey` (= `order_id`), оба адаптера обязаны гарантировать, что повторный вызов с тем же ключом не создаёт дублирующий платёж (Mock — по таблице `payment.payments` unique(order_id); YooKassa — через их встроенный `Idempotence-Key` HTTP-заголовок).

## Альтернативы

| Вариант | Почему отклонён |
|---|---|
| Только Mock | Не показывает умение интегрироваться с реальным внешним платёжным API — терялась часть портфолио-ценности |
| Только YooKassa | Усложняет demo/CI (требует секретов), недоступно ревьюеру без своего sandbox-аккаунта |
| Stripe вместо YooKassa | YooKassa более релевантен для RU-контекста (пользователь base в Ростове-на-Дону) и типичен для рынка РФ, где работает Head of Development |

## Последствия

- (+) Домен `order`/`payment` не знает, какой провайдер используется — переключение — конфигурация, не код.
- (+) CI и локальный запуск по умолчанию работают offline (Mock), реальная интеграция — опциональная демонстрация.
- (-) Два адаптера = два набора edge-кейсов для тестирования (обе реализации нужно покрыть contract-тестами против одного интерфейса — Epic PAYMENT включает "provider contract test suite").
