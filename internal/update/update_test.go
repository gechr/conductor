package update_test

import (
	"testing"

	"github.com/gechr/conductor/internal/update"
	"github.com/stretchr/testify/require"
)

func TestRequested(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		requested bool
		err       string
	}{
		{name: "alone", args: []string{"+update"}, requested: true},
		{name: "absent", args: []string{"build", "--verbose"}},
		{name: "empty"},
		{
			name:      "with other arguments",
			args:      []string{"--verbose", "+update"},
			requested: true,
			err:       "+update cannot be combined with other arguments",
		},
		{name: "after terminator", args: []string{"--", "+update"}},
		{name: "old flag spelling", args: []string{"--self-update"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requested, err := update.Requested(tt.args)
			require.Equal(t, tt.requested, requested)
			if tt.err == "" {
				require.NoError(t, err)
				return
			}
			require.EqualError(t, err, tt.err)
		})
	}
}
