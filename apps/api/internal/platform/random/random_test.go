package random

import "testing"

func TestFloat64Range_WithinBounds(t *testing.T) {
	src := New(42)
	for i := 0; i < 1000; i++ {
		v := src.Float64Range(100, 500)
		if v < 100 || v >= 500 {
			t.Fatalf("Float64Range(100,500) produced out-of-range value: %v", v)
		}
	}
}

func TestIntRange_WithinBoundsInclusive(t *testing.T) {
	src := New(42)
	seenMin, seenMax := false, false
	for i := 0; i < 5000; i++ {
		v := src.IntRange(1, 5)
		if v < 1 || v > 5 {
			t.Fatalf("IntRange(1,5) produced out-of-range value: %d", v)
		}
		if v == 1 {
			seenMin = true
		}
		if v == 5 {
			seenMax = true
		}
	}
	if !seenMin || !seenMax {
		t.Errorf("expected to see both boundary values 1 and 5 over 5000 draws, seenMin=%v seenMax=%v", seenMin, seenMax)
	}
}

func TestBool_ApproximatesProbability(t *testing.T) {
	src := New(42)
	trueCount := 0
	const n = 10000
	for i := 0; i < n; i++ {
		if src.Bool(0.1) {
			trueCount++
		}
	}
	ratio := float64(trueCount) / float64(n)
	// Допускаем статистический разброс +/- 3 процентных пункта вокруг 10%.
	if ratio < 0.07 || ratio > 0.13 {
		t.Errorf("expected Bool(0.1) true-ratio near 0.10 over %d draws, got %.3f", n, ratio)
	}
}

func TestPick_ReturnsOnlyProvidedItems(t *testing.T) {
	src := New(42)
	items := []string{"a", "b", "c"}
	seen := map[string]bool{}
	for i := 0; i < 200; i++ {
		v := src.Pick(items)
		if v != "a" && v != "b" && v != "c" {
			t.Fatalf("Pick returned unexpected value: %q", v)
		}
		seen[v] = true
	}
	if len(seen) != 3 {
		t.Errorf("expected to see all 3 items over 200 draws, saw %d distinct values", len(seen))
	}
}

func TestPick_PanicsOnEmptySlice(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on empty slice, got none")
		}
	}()
	src := New(42)
	src.Pick([]string{})
}

func TestNew_SameSeedProducesSameSequence(t *testing.T) {
	a := New(7)
	b := New(7)
	for i := 0; i < 20; i++ {
		va := a.Float64Range(0, 1)
		vb := b.Float64Range(0, 1)
		if va != vb {
			t.Fatalf("expected deterministic sequence for same seed, diverged at draw %d: %v != %v", i, va, vb)
		}
	}
}
