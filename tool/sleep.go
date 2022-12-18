package tools

import (
	"context"
	"time"
)

// Sleep is similar to time.Sleep, but it can be interrupted by a ctx.Done() closing.
func Sleep(ctx context.Context, d time.Duration) {
	t := time.NewTimer(d)
	select {
	case <-ctx.Done():
		t.Stop()
		return
	case <-t.C:
		return
	}
}

// SleepStopChan is similar to time.Sleep, but it can be interrupted by a channel closing.
func SleepStopChan(ch chan struct{}, d time.Duration) {
	t := time.NewTimer(d)
	select {
	case <-ch:
		t.Stop()
		return
	case <-t.C:
		return
	}
}
