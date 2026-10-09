package perky_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/alexdrel/perky"
	"github.com/alexdrel/perky/examples/go/environment"
	"github.com/alexdrel/perky/examples/go/logger"
)

func TestLoggerWithLocalClockAndOutput(t *testing.T) {
	var lines []string
	ctx := perky.New(
		environment.Now.Bind(time.Unix(1, 0)),
		environment.WriteLine.Bind(func(line string) { lines = append(lines, line) }),
		logger.Colors.Bind(false),
		logger.LogLevel.Bind(logger.Debug),
	)

	logger.Log(ctx, logger.Debug, "Connecting")
	quiet := ctx.With(logger.LogLevel.Bind(logger.Warn))
	logger.Log(quiet, logger.Info, "Hidden")
	logger.Log(quiet, logger.Warn, "Retrying")
	logger.Log(ctx, logger.Info, "Still visible")
	logger.Log(ctx.With(logger.Colors.Bind(true)), logger.Warn, "Colored warning")

	want := []string{
		"1970-01-01T00:00:01.000Z DEBUG Connecting",
		"1970-01-01T00:00:01.000Z WARN Retrying",
		"1970-01-01T00:00:01.000Z INFO Still visible",
		"1970-01-01T00:00:01.000Z \x1b[33mWARN\x1b[0m Colored warning",
	}
	if !reflect.DeepEqual(lines, want) {
		t.Fatalf("log output:\n got: %q\nwant: %q", lines, want)
	}
}
