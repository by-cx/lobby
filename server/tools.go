package server

import "github.com/google/go-cmp/cmp"

// Compare compares discovery A and B and returns true if those two are different.
func Compare(discoveryA, discoveryB Discovery) bool {
	discoveryA.LastCheck = 0
	discoveryB.LastCheck = 0
	discoveryA.TTL = 0
	discoveryB.TTL = 0
	return !cmp.Equal(discoveryA, discoveryB)
}
