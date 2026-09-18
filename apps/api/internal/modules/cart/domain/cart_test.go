package domain_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/ostkost/dopamine-market/api/internal/modules/cart/domain"
)

func TestCart_ImmutabilityAndOperations(t *testing.T) {
	c := domain.NewCart(nil, nil)
	if !c.IsEmpty() {
		t.Fatalf("expected empty cart")
	}

	prod1 := uuid.New()
	prod2 := uuid.New()
	pickup := uuid.New()

	// 1. WithItem returns new copy, original unchanged
	c1, err := c.WithItem(prod1, 2)
	if err != nil {
		t.Fatalf("WithItem failed: %v", err)
	}
	if !c.IsEmpty() {
		t.Fatalf("original cart was mutated!")
	}
	if c1.TotalQuantity() != 2 {
		t.Errorf("expected total quantity 2, got %d", c1.TotalQuantity())
	}

	// 2. Add second item
	c2, err := c1.WithItem(prod2, 3)
	if err != nil {
		t.Fatalf("WithItem failed: %v", err)
	}
	if c2.TotalQuantity() != 5 {
		t.Errorf("expected total quantity 5, got %d", c2.TotalQuantity())
	}

	// 3. WithPickupPoint
	c3 := c2.WithPickupPoint(&pickup)
	if c2.PickupPointID() != nil {
		t.Errorf("c2 pickup point was mutated!")
	}
	if c3.PickupPointID() == nil || *c3.PickupPointID() != pickup {
		t.Errorf("expected pickup point %v, got %v", pickup, c3.PickupPointID())
	}

	// 4. Update quantity
	c4, err := c3.WithQuantity(prod1, 5)
	if err != nil {
		t.Fatalf("WithQuantity failed: %v", err)
	}
	if c4.Items()[prod1] != 5 {
		t.Errorf("expected prod1 quantity 5, got %d", c4.Items()[prod1])
	}

	// 5. WithoutItem
	c5 := c4.WithoutItem(prod1)
	if _, ok := c5.Items()[prod1]; ok {
		t.Errorf("expected prod1 to be removed")
	}
	if c5.TotalQuantity() != 3 {
		t.Errorf("expected remaining quantity 3, got %d", c5.TotalQuantity())
	}
	// Pickup point is preserved
	if c5.PickupPointID() == nil || *c5.PickupPointID() != pickup {
		t.Errorf("pickup point lost after removing item")
	}
}
