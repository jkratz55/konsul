package zerolog

import (
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

func TestWrapper(t *testing.T) {
	logger := zerolog.New(os.Stdout)
	wrapper := Wrap(logger)
	wrapper.Info("oh no something went wrong",
		"err", nil,
		"duruation", time.Second*10,
		"datetime", time.Now(),
		"counter", 199,
		"dada", 3242343)
}
