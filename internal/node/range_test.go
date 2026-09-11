package node

import "testing"

func TestResolveEnginesRange(t *testing.T) {
	cases := map[string]string{
		"":         DefaultNodeVersion,
		"*":        DefaultNodeVersion,
		">=18":     "24",
		">=18.0.0": "24",
		">=22.0.0": "24",
		"20.x":     "20",
		"20":       "20",
		"^20.15.0": "20",
		"~18.20.0": "18",
		">=20 <22": "20",
		">=16 <18": "16",
		"abc":      DefaultNodeVersion,
	}
	for input, want := range cases {
		got := ResolveEnginesRange(input)
		if got != want {
			t.Errorf("ResolveEnginesRange(%q)=%s want %s", input, got, want)
		}
	}
}
