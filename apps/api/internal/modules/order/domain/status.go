package domain

// Status представляет статус жизненного цикла заказа.
type Status string

const (
	StatusCreated         Status = "created"
	StatusPaymentPending  Status = "payment_pending"
	StatusPaid            Status = "paid"
	StatusAssembling      Status = "assembling"
	StatusCourierAssigned Status = "courier_assigned"
	StatusInTransit       Status = "in_transit"
	StatusDelivered       Status = "delivered"
	StatusPaymentFailed   Status = "payment_failed"
	StatusCancelled       Status = "cancelled"
)

func (s Status) String() string {
	return string(s)
}

// IsTerminal проверяет, является ли статус конечным.
func (s Status) IsTerminal() bool {
	return s == StatusDelivered || s == StatusCancelled
}

// allowedTransitions определяет допустимые переходы стейт-машины заказа (FR-ORDER-02).
//
// Mermaid State Machine Diagram:
//
// ```mermaid
// stateDiagram-v2
//     [*] --> created
//     created --> payment_pending
//     created --> cancelled
//     payment_pending --> paid
//     payment_pending --> payment_failed
//     payment_pending --> cancelled
//     payment_failed --> payment_pending
//     payment_failed --> cancelled
//     paid --> assembling
//     paid --> cancelled
//     assembling --> courier_assigned
//     assembling --> cancelled
//     courier_assigned --> in_transit
//     courier_assigned --> cancelled
//     in_transit --> delivered
//     in_transit --> cancelled
//     delivered --> [*]
//     cancelled --> [*]
// ```
var allowedTransitions = map[Status][]Status{
	StatusCreated: {
		StatusPaymentPending,
		StatusCancelled,
	},
	StatusPaymentPending: {
		StatusPaid,
		StatusPaymentFailed,
		StatusCancelled,
	},
	StatusPaymentFailed: {
		StatusPaymentPending,
		StatusCancelled,
	},
	StatusPaid: {
		StatusAssembling,
		StatusCancelled,
	},
	StatusAssembling: {
		StatusCourierAssigned,
		StatusCancelled,
	},
	StatusCourierAssigned: {
		StatusInTransit,
		StatusCancelled,
	},
	StatusInTransit: {
		StatusDelivered,
		StatusCancelled,
	},
	StatusDelivered: {},
	StatusCancelled: {},
}

// CanTransitionTo проверяет допустимость перехода из текущего статуса в целевой.
func (s Status) CanTransitionTo(target Status) bool {
	validTargets, ok := allowedTransitions[s]
	if !ok {
		return false
	}
	for _, valid := range validTargets {
		if valid == target {
			return true
		}
	}
	return false
}
