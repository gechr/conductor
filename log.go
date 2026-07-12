package conductor

import (
	"time"

	"github.com/gechr/clog"
	"github.com/gechr/x/human"
)

// defaultTimeGradientMax is the duration mapped to the end of the duration
// and elapsed gradients (green at zero, red at the maximum). Override it in
// [App.ConfigureLog] via [clog.SetTimeGradientMax] - or the per-field
// [clog.SetDurationGradientMax] and [clog.SetElapsedGradientMax] - and set 0
// to disable the gradient.
const defaultTimeGradientMax = 20 * time.Second

// LogDefaults applies the house clog style shared by every tool: stderr
// output with automatic color detection, space-separated slices,
// human-friendly duration fields, and gradient-colored duration and elapsed
// values. It is exported so a tool that skips Conductor entirely can still
// opt into the house style.
func LogDefaults() {
	clog.SetOutput(clog.Stderr(clog.ColorAuto))
	clog.SetSliceSeparator(" ")

	// Start from the current formats so env-loaded settings survive.
	formats := clog.Default.FieldFormats()
	formats.DurationFormat = human.FormatDuration
	clog.SetFieldFormats(formats)
	clog.SetTimeGradientMax(defaultTimeGradientMax)
}
