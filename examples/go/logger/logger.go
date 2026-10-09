// Package logger owns logging policy and reads application effects from its context.
package logger

import (
	"fmt"

	"github.com/alexdrel/perky"
	"github.com/alexdrel/perky/examples/go/environment"
)

type Level int

const (
	Debug Level = iota
	Info
	Warn
)

var LogLevel = perky.Key(Info)
var Colors = perky.Key(true)

var labels = [...]string{"DEBUG", "INFO", "WARN"}
var colors = [...]string{"\x1b[90m", "\x1b[36m", "\x1b[33m"}

// Log accepts either native Go contexts or standalone Perky contexts.
func Log(ctx perky.Values, level Level, message string) {
	if level < LogLevel.Get(ctx) {
		return
	}
	timestamp := environment.Now.Get(ctx).UTC().Format("2006-01-02T15:04:05.000Z")
	label := labels[level]
	if Colors.Get(ctx) {
		label = colors[level] + label + "\x1b[0m"
	}
	environment.WriteLine.Get(ctx)(fmt.Sprintf("%s %s %s", timestamp, label, message))
}
