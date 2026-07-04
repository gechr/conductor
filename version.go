package conductor

import "github.com/gechr/clive"

// Version returns the running binary's version via [clive.Current]:
// ldflag-injected, module build info, or VCS revision - whichever resolves
// first.
func (r *Runtime) Version() string {
	return clive.Current()
}

// PrintVersion writes the version to stdout: the bare string, or a labelled
// table of build details (with release hyperlinks when Module/Repo is set)
// when detailed is true.
func (r *Runtime) PrintVersion(detailed bool) {
	if detailed {
		r.Info.PrintDetailed()
		return
	}
	clive.Print()
}
