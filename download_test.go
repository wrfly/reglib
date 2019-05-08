package reglib

import (
	"testing"
)

func TestSplitRanges(t *testing.T) {
	x := splitRanges(int(30*mbSize) + 1)
	t.Log(x)
}
