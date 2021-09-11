package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompare(t *testing.T) {
	discoveryA := Discovery{
		Hostname: "abcd.com",
		Labels: Labels{
			Label("label1"),
		},
		LastCheck: 52,
	}
	discoveryAB := Discovery{
		Hostname: "abcd.com",
		Labels: Labels{
			Label("label1"),
		},
		LastCheck: 56,
	}
	discoveryB := Discovery{
		Hostname: "efgh.com",
		Labels: Labels{
			Label("label2"),
		},
		LastCheck: 56,
	}
	discoveryC := Discovery{
		Hostname: "abcd.com",
		Labels: Labels{
			Label("label2"),
		},
		LastCheck: 60,
	}

	assert.True(t, Compare(discoveryA, discoveryB))
	assert.True(t, Compare(discoveryB, discoveryC))
	assert.True(t, Compare(discoveryA, discoveryC)) // Test different labels and same hostname
	assert.False(t, Compare(discoveryA, discoveryA))
	assert.False(t, Compare(discoveryB, discoveryB))
	assert.False(t, Compare(discoveryA, discoveryAB)) // Test that last check is zeroed
}
