package testlogger

import (
	"context"
	"io"
	"os"

	"github.com/obalunenko/getenv"
	log "github.com/obalunenko/logger"
)

const (
	discardEnv = "TEST_DISCARD_LOG"
)

type noopCloser struct {
	io.Writer
}

func (noopCloser) Close() error {
	return nil
}

// New returns context with logger.
// If TEST_DISCARD_LOG env var is set to true, logger will be discarded.
func New(ctx context.Context) context.Context {
	w := io.Discard

	if !getenv.EnvOrDefault(discardEnv, false) {
		w = os.Stderr
	}

	l := log.Init(ctx, log.Params{
		Writer: noopCloser{Writer: w},
		Level:  "debug",
		Format: "text",
	})

	return log.ContextWithLogger(ctx, l)
}
