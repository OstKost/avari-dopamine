package domain_test

import (
	"testing"

	"github.com/ostkost/dopamine-market/api/internal/modules/order/domain"
)

func TestStatus_Transitions(t *testing.T) {
	allStatuses := []domain.Status{
		domain.StatusCreated,
		domain.StatusPaymentPending,
		domain.StatusPaid,
		domain.StatusAssembling,
		domain.StatusCourierAssigned,
		domain.StatusInTransit,
		domain.StatusDelivered,
		domain.StatusPaymentFailed,
		domain.StatusCancelled,
	}

	validTransitions := map[domain.Status][]domain.Status{
		domain.StatusCreated: {
			domain.StatusPaymentPending,
			domain.StatusCancelled,
		},
		domain.StatusPaymentPending: {
			domain.StatusPaid,
			domain.StatusPaymentFailed,
			domain.StatusCancelled,
		},
		domain.StatusPaymentFailed: {
			domain.StatusPaymentPending,
			domain.StatusCancelled,
		},
		domain.StatusPaid: {
			domain.StatusAssembling,
			domain.StatusCancelled,
		},
		domain.StatusAssembling: {
			domain.StatusCourierAssigned,
			domain.StatusCancelled,
		},
		domain.StatusCourierAssigned: {
			domain.StatusInTransit,
			domain.StatusCancelled,
		},
		domain.StatusInTransit: {
			domain.StatusDelivered,
			domain.StatusCancelled,
		},
		domain.StatusDelivered: {},
		domain.StatusCancelled: {},
	}

	for _, from := range allStatuses {
		expectedValids := validTransitions[from]
		for _, to := range allStatuses {
			expectedValid := false
			for _, v := range expectedValids {
				if v == to {
					expectedValid = true
					break
				}
			}

			can := from.CanTransitionTo(to)
			if can != expectedValid {
				t.Errorf("transition %s -> %s = %v, expected %v", from, to, can, expectedValid)
			}
		}
	}
}

func TestStatus_Terminal(t *testing.T) {
	if !domain.StatusDelivered.IsTerminal() {
		t.Errorf("expected delivered to be terminal")
	}
	if !domain.StatusCancelled.IsTerminal() {
		t.Errorf("expected cancelled to be terminal")
	}
	if domain.StatusInTransit.IsTerminal() {
		t.Errorf("in_transit should not be terminal")
	}
}
