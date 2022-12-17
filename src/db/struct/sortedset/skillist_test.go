package sortedset

import "testing"

func TestGet_FirstInScope(t *testing.T) {
	var s = makeSkipList()

	s.insert("m1", 0)
	s.insert("m2", 1)
	s.insert("m3", 2)
	s.insert("m4", 3)
	s.insert("m5", 4)
	var node = s.getFirstInScoreRange(&ScoreBorder{
		Value:   1,
		Exclude: false,
	}, &ScoreBorder{
		Value:   3,
		Exclude: true,
	})
	if node == nil || node.Member != "m2" {
		t.Fail()
	}
}
