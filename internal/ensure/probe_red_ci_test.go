package ensure

import "testing"

// Deliberate failure used once to prove the default branch's required `ci`
// check blocks an admin merge. Deleted with its branch.
func TestProbeRedCI(t *testing.T) {
	if 1 != 2 {
		t.Fatal("probe: required ci must be red")
	}
}
