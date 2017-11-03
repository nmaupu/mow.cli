package fsm

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSimplify(t *testing.T) {

	cases := []struct {
		original   string
		simplified string
	}{

		{
			original: `
					S1 * S2
					S2 -a (S3)
			`,
			simplified: "S1 -a (S3)",
		},
		{
			// seq like FSM
			original: `
					S1 *  S2
					S2 -a S3
					S3 *  S2
					S3 *  (S4)
			`,
			simplified: `
					S1 -a (S3)
					(S3) -a (S3)
			`,
		},
		{
			// optional transition FSM
			original: `
					S1 -a  S2
					S2 * (S3)
					S1 *  (S3)
			`,
			simplified: `
					(S1) -a (S2)
			`,
		},
	}

	for _, cas := range cases {
		t.Logf("FSM.simplify:\noriginal: %s\nexpected: %s", cas.original, cas.simplified)

		original := mkFsm(cas.original, testMatchers)

		original.Prepare()

		simplified := mkFsm(cas.simplified, testMatchers)

		require.Equal(t, fsmStr(simplified), fsmStr(original))
	}
}

func TestSort(t *testing.T) {
	s := mkFsm(`
		S1 * S2
		S1 * S3
		S1 ARG S2
		S1 -- S2
		S1 [OPTIONS] S2
		S1 -a S3
	`, testMatchers)

	s.Prepare()

	require.Equal(t, []string{"-a", "[OPTIONS]", "ARG", "--"}, mkTransitionStrs(s.Transitions))
}
