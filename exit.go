package conductor

import (
	"errors"

	"github.com/gechr/clive/updater"
	"github.com/gechr/clog"
)

// ExitCode maps a command error to a process exit code, shared by all
// framework adapters: nil is 0; [updater.ErrReported] is 1 without further
// output (the failure is already on screen); otherwise the optional custom
// mapper decides, falling back to logging the error and returning 1.
func ExitCode(err error, custom func(error) int) int {
	if err == nil {
		return 0
	}
	if errors.Is(err, updater.ErrReported) {
		return 1
	}
	if custom != nil {
		return custom(err)
	}
	clog.Error().Err(err).Msg("Failed")
	return 1
}
