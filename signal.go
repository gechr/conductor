package conductor

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// SignalContext returns a context cancelled on SIGINT or SIGTERM, for
// commands that block on external interaction (network calls, hardware
// prompts) and should abort cleanly on Ctrl-C. Opt-in: create it in main and
// thread it through the framework adapter's Run or command bindings. The stop
// function releases the signal registration; defer it.
func SignalContext() (context.Context, context.CancelFunc) {
	return signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
}
