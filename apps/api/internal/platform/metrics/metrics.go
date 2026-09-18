package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	OrdersCreatedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "dopamine",
			Subsystem: "orders",
			Name:      "created_total",
			Help:      "Общее количество созданных заказов",
		},
	)

	OrdersPaidTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "dopamine",
			Subsystem: "orders",
			Name:      "paid_total",
			Help:      "Общее количество успешно оплаченных заказов",
		},
	)

	PaymentsInitiatedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "dopamine",
			Subsystem: "payments",
			Name:      "initiated_total",
			Help:      "Общее количество инициированных платежей по провайдерам",
		},
		[]string{"provider"},
	)

	PaymentsSucceededTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "dopamine",
			Subsystem: "payments",
			Name:      "succeeded_total",
			Help:      "Общее количество успешных платежей",
		},
		[]string{"provider"},
	)

	PaymentsFailedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "dopamine",
			Subsystem: "payments",
			Name:      "failed_total",
			Help:      "Общее количество неудавшихся платежей",
		},
		[]string{"provider"},
	)

	HttpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "dopamine",
			Subsystem: "http",
			Name:      "request_duration_seconds",
			Help:      "Длительность HTTP запросов в секундах",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)
)

func init() {
	prometheus.MustRegister(
		OrdersCreatedTotal,
		OrdersPaidTotal,
		PaymentsInitiatedTotal,
		PaymentsSucceededTotal,
		PaymentsFailedTotal,
		HttpRequestDuration,
	)
}

// Handler возвращает HTTP хендлер для /metrics эндпоинта.
func Handler() http.Handler {
	return promhttp.Handler()
}
