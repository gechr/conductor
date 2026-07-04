package conductor

import (
	"time"

	"github.com/gechr/clog"
	"github.com/gechr/x/human"
)

// gradientMax is the duration mapped to the end of the duration and elapsed
// field gradients.
const gradientMax = 20 * time.Second

// LogDefaults applies the house clog style shared by every tool: stderr
// output with automatic color detection, space-separated slices, and
// human-friendly duration fields with a 20s gradient. It is exported so a
// tool that skips conductor entirely can still opt into the house style.
func LogDefaults() {
	clog.SetOutput(clog.Stderr(clog.ColorAuto))
	clog.SetSliceSeparator(" ")

	// Start from the current formats so env-loaded settings survive.
	formats := clog.Default.FieldFormats()
	formats.DurationFormat = human.FormatDuration
	formats.DurationGradientMax = gradientMax
	formats.ElapsedGradientMax = gradientMax
	clog.SetFieldFormats(formats)
}
