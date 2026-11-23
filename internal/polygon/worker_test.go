package polygon

import (
	"context"
	"testing"
	"time"

	"stonks/internal/models"
)

func TestWaitForWorkerCycle(t *testing.T) {
	t.Run("interval elapsed", func(t *testing.T) {
		ctx := context.Background()
		if !waitForWorkerCycle(ctx, time.Millisecond) {
			t.Fatal("waitForWorkerCycle returned false before interval elapsed")
		}
	})

	t.Run("context cancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		started := time.Now()
		if waitForWorkerCycle(ctx, time.Hour) {
			t.Fatal("waitForWorkerCycle returned true after cancellation")
		}
		if elapsed := time.Since(started); elapsed > 100*time.Millisecond {
			t.Fatalf("cancellation took %s", elapsed)
		}
	})
}

func TestBuildPolygonOptionTicker(t *testing.T) {
	expiration := time.Date(2027, time.January, 16, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name string
		opt  *models.Option
		want string
	}{
		{
			name: "call",
			opt:  &models.Option{Symbol: "aapl", Type: "Call", Strike: 150, Expiration: expiration},
			want: "O:AAPL270116C00150000",
		},
		{
			name: "fractional strike rounds",
			opt:  &models.Option{Symbol: "spy", Type: "Put", Strike: 123.4567, Expiration: expiration},
			want: "O:SPY270116P00123457",
		},
		{
			name: "unknown type",
			opt:  &models.Option{Symbol: "SPY", Type: "Unknown", Strike: 100, Expiration: expiration},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildPolygonOptionTicker(tt.opt); got != tt.want {
				t.Fatalf("buildPolygonOptionTicker() = %q, want %q", got, tt.want)
			}
		})
	}
}
