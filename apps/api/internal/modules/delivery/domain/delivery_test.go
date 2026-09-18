package domain_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ostkost/dopamine-market/api/internal/modules/delivery/domain"
)

func TestDelivery_Transitions(t *testing.T) {
	t.Parallel()

	t.Run("valid full happy path", func(t *testing.T) {
		del := domain.NewDelivery(uuid.New(), uuid.New(), "Алексей", 4.9, time.Now(), 5*time.Minute)

		if del.Status != domain.StatusAssembling {
			t.Fatalf("expected initial status %q, got %q", domain.StatusAssembling, del.Status)
		}

		if err := del.TransitionTo(domain.StatusCourierAssigned); err != nil {
			t.Fatalf("unexpected error on courier_assigned transition: %v", err)
		}
		if del.Status != domain.StatusCourierAssigned {
			t.Errorf("expected status %q, got %q", domain.StatusCourierAssigned, del.Status)
		}

		if err := del.TransitionTo(domain.StatusInTransit); err != nil {
			t.Fatalf("unexpected error on in_transit transition: %v", err)
		}
		if del.Status != domain.StatusInTransit {
			t.Errorf("expected status %q, got %q", domain.StatusInTransit, del.Status)
		}

		if err := del.TransitionTo(domain.StatusDelivered); err != nil {
			t.Fatalf("unexpected error on delivered transition: %v", err)
		}
		if del.Status != domain.StatusDelivered {
			t.Errorf("expected status %q, got %q", domain.StatusDelivered, del.Status)
		}
		if !del.IsTerminal() {
			t.Errorf("expected terminal state")
		}
	})

	t.Run("valid delayed branch", func(t *testing.T) {
		del := domain.NewDelivery(uuid.New(), uuid.New(), "Михаил", 4.7, time.Now(), 5*time.Minute)
		_ = del.TransitionTo(domain.StatusCourierAssigned)
		_ = del.TransitionTo(domain.StatusInTransit)

		if err := del.TransitionTo(domain.StatusDeliveryDelayed); err != nil {
			t.Fatalf("unexpected error on delivery_delayed transition: %v", err)
		}
		if del.Status != domain.StatusDeliveryDelayed {
			t.Errorf("expected status %q, got %q", domain.StatusDeliveryDelayed, del.Status)
		}

		if err := del.TransitionTo(domain.StatusDelivered); err != nil {
			t.Fatalf("unexpected error on delivered after delayed: %v", err)
		}
		if del.Status != domain.StatusDelivered {
			t.Errorf("expected status %q, got %q", domain.StatusDelivered, del.Status)
		}
	})

	t.Run("invalid skip transition", func(t *testing.T) {
		del := domain.NewDelivery(uuid.New(), uuid.New(), "Иван", 4.5, time.Now(), 5*time.Minute)

		// Cannot skip from assembling directly to in_transit
		if err := del.TransitionTo(domain.StatusInTransit); err == nil {
			t.Errorf("expected error transitioning from assembling directly to in_transit, got nil")
		}

		// Cannot skip from assembling directly to delivered
		if err := del.TransitionTo(domain.StatusDelivered); err == nil {
			t.Errorf("expected error transitioning from assembling directly to delivered, got nil")
		}
	})

	t.Run("no transitions from delivered", func(t *testing.T) {
		del := domain.NewDelivery(uuid.New(), uuid.New(), "Дмитрий", 4.8, time.Now(), 5*time.Minute)
		_ = del.TransitionTo(domain.StatusCourierAssigned)
		_ = del.TransitionTo(domain.StatusInTransit)
		_ = del.TransitionTo(domain.StatusDelivered)

		if err := del.TransitionTo(domain.StatusInTransit); err == nil {
			t.Errorf("expected error transitioning backwards from delivered, got nil")
		}
	})
}
