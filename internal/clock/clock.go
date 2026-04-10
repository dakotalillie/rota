package clock

import (
	"os"
	"strings"
	"time"

	"github.com/dakotalillie/rota/internal/domain"
)

// Clock uses time.Now().
type Clock struct{}

func New() *Clock { return &Clock{} }

func (c *Clock) Now() time.Time { return time.Now() }

// FSClock reads the current time from a file on each call.
// The file must contain an RFC3339 timestamp.
// Falls back to time.Now() if the file doesn't exist, is empty, or has invalid content.
type FSClock struct {
	path string
}

func NewFS(path string) *FSClock { return &FSClock{path: path} }

func (c *FSClock) Now() time.Time {
	data, err := os.ReadFile(c.path)
	if err != nil {
		return time.Now()
	}
	s := strings.TrimSpace(string(data))
	if s == "" {
		return time.Now()
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Now()
	}
	return t
}

var _ domain.Clock = (*Clock)(nil)
var _ domain.Clock = (*FSClock)(nil)
