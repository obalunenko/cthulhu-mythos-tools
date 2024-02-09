package config

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/obalunenko/cthulhu-mythos-tools/internal/testlogger"
)

func unsetEnv(tb testing.TB) {
	tb.Helper()

	tb.Setenv(portEnv, "")
	tb.Setenv(hostEnv, "")
	tb.Setenv(levelEnv, "")
	tb.Setenv(formatEnv, "")
}

func TestLoadDefault(t *testing.T) {
	ctx := testlogger.New(context.Background())

	unsetEnv(t)

	t.Run("default", func(t *testing.T) {
		cfg, err := Load(ctx)
		require.NoError(t, err)

		require.Equal(t, DefaultConfig(), cfg)
	})

	t.Run("env", func(t *testing.T) {
		t.Run("port", func(t *testing.T) {
			t.Setenv(portEnv, "8081")

			cfg, err := Load(ctx)
			require.NoError(t, err)

			expected := DefaultConfig()
			expected.HTTP.Port = "8081"

			assert.Equal(t, expected, cfg)
		})
		t.Run("host", func(t *testing.T) {
			t.Setenv(hostEnv, "127.0.0.1")

			cfg, err := Load(ctx)
			require.NoError(t, err)

			expected := DefaultConfig()
			expected.HTTP.Host = "127.0.0.1"

			assert.Equal(t, expected, cfg)
		})
		t.Run("level", func(t *testing.T) {
			t.Setenv(levelEnv, "DEBUG")

			cfg, err := Load(ctx)
			require.NoError(t, err)

			expected := DefaultConfig()
			expected.Log.Level = "DEBUG"

			assert.Equal(t, expected, cfg)
		})
		t.Run("format", func(t *testing.T) {
			t.Setenv(formatEnv, "json")

			cfg, err := Load(ctx)
			require.NoError(t, err)

			expected := DefaultConfig()
			expected.Log.Format = "json"

			assert.Equal(t, expected, cfg)
		})
	})
}
