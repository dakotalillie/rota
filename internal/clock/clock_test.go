package clock_test

import (
	"os"
	"testing"
	"time"

	"github.com/dakotalillie/rota/internal/clock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClock_Now(t *testing.T) {
	clk := clock.New()
	before := time.Now()
	got := clk.Now()
	after := time.Now()
	assert.False(t, got.Before(before))
	assert.False(t, got.After(after))
}

func TestFSClock_Now_readsFile(t *testing.T) {
	f, err := os.CreateTemp("", "fsclock-test-*.txt")
	require.NoError(t, err)
	defer os.Remove(f.Name()) //nolint:errcheck

	want := time.Date(2026, 4, 9, 12, 0, 0, 0, time.UTC)
	_, err = f.WriteString(want.Format(time.RFC3339))
	require.NoError(t, err)
	f.Close() //nolint:errcheck

	clk := clock.NewFS(f.Name())
	got := clk.Now()
	assert.Equal(t, want, got)
}

func TestFSClock_Now_fallbackWhenMissing(t *testing.T) {
	clk := clock.NewFS("/nonexistent/path/does-not-exist-clock.txt")
	before := time.Now()
	got := clk.Now()
	after := time.Now()
	assert.False(t, got.Before(before))
	assert.False(t, got.After(after))
}

func TestFSClock_Now_fallbackWhenInvalid(t *testing.T) {
	f, err := os.CreateTemp("", "fsclock-test-*.txt")
	require.NoError(t, err)
	defer os.Remove(f.Name()) //nolint:errcheck

	_, err = f.WriteString("not-a-valid-time")
	require.NoError(t, err)
	f.Close() //nolint:errcheck

	clk := clock.NewFS(f.Name())
	before := time.Now()
	got := clk.Now()
	after := time.Now()
	assert.False(t, got.Before(before))
	assert.False(t, got.After(after))
}
