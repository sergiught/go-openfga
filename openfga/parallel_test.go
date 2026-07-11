package openfga

import (
	"context"
	"sync/atomic"
	"testing"
)

func TestRunParallel_RunsAllAndBoundsConcurrency(t *testing.T) {
	const n = 20
	var inFlight, maxSeen int32
	errs := runParallel(context.Background(), n, 3, func(i int) error {
		cur := atomic.AddInt32(&inFlight, 1)
		for {
			old := atomic.LoadInt32(&maxSeen)
			if cur <= old || atomic.CompareAndSwapInt32(&maxSeen, old, cur) {
				break
			}
		}
		atomic.AddInt32(&inFlight, -1)
		return nil
	})
	if len(errs) != n {
		t.Fatalf("len(errs) = %d, want %d", len(errs), n)
	}
	if maxSeen > 3 {
		t.Fatalf("max concurrency = %d, want <= 3", maxSeen)
	}
}

func TestRunParallel_CollectsPerIndexErrors(t *testing.T) {
	errs := runParallel(context.Background(), 4, 2, func(i int) error {
		if i == 2 {
			return context.Canceled
		}
		return nil
	})
	if errs[2] == nil || errs[0] != nil {
		t.Fatalf("errs = %v, want only index 2 non-nil", errs)
	}
}

func TestRunParallel_StopsStartingAfterCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	errs := runParallel(ctx, 5, 2, func(i int) error { return nil })
	for i, e := range errs {
		if e == nil {
			t.Fatalf("index %d ran despite cancelled ctx", i)
		}
	}
}

func TestResolvePositive(t *testing.T) {
	if resolvePositive(0, 10) != 10 || resolvePositive(-1, 10) != 10 || resolvePositive(5, 10) != 5 {
		t.Fatal("resolvePositive wrong")
	}
}
