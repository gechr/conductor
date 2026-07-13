package conductor

import (
	"slices"
	"strings"

	"github.com/gechr/clive"
	"github.com/gechr/clive/notify"
	xslices "github.com/gechr/x/slices"
)

// notifySkipVerbs are command verbs that never trigger an update check: the
// update flow itself, version output, help, and shell-completion machinery.
var notifySkipVerbs = []string{"update", "version", "help", "completion", "__complete"}

// Notify starts the passive update check for the resolved command and
// returns a flush function - never nil - that the caller runs after the
// command completes, so the hint trails normal output. The check is skipped
// when App.Updater is nil, when the command's verb is one of the built-in
// skip verbs or App.NotifySkip, or when App.NotifyOnly is set and does not
// contain the verb. clive's <APP>_NO_UPDATE_CHECK kill switch still
// applies underneath.
func (r *Runtime) Notify(command string) func() {
	noop := func() {}
	if r.App.Updater == nil {
		return noop
	}
	if r.notifySkipped(command) {
		return noop
	}
	opts := append(
		[]notify.Option{notify.WithCurrentVersion(clive.Current())},
		r.App.NotifyOptions...,
	)
	return notify.Check(r.App.Updater, opts...)
}

// notifySkipped reports whether the update check is suppressed for the
// command's leading verb.
func (r *Runtime) notifySkipped(command string) bool {
	verb, _, _ := strings.Cut(strings.TrimSpace(command), " ")
	if len(r.App.NotifyOnly) > 0 {
		return !slices.Contains(r.App.NotifyOnly, verb)
	}
	return xslices.ContainedByAny(verb, notifySkipVerbs, r.App.NotifySkip)
}
