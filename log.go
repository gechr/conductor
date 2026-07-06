package conductor

import (
	"github.com/gechr/clog"
	"github.com/gechr/x/human"
)

// LogDefaults applies the house clog style shared by every tool: stderr
// output with automatic color detection, space-separated slices, and
// human-friendly duration fields. It is exported so a tool that skips
// Conductor entirely can still opt into the house style.
func LogDefaults() {
	clog.SetOutput(clog.Stderr(clog.ColorAuto))
	clog.SetSliceSeparator(" ")

	// Start from the current formats so env-loaded settings survive.
	formats := clog.Default.FieldFormats()
	formats.DurationFormat = human.FormatDuration
	clog.SetFieldFormats(formats)
}
