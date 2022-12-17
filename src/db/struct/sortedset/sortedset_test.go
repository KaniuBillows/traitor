package sortedset

import "testing"

func TestSortedSet_PopMin(t *testing.T) {
	var set = Make()
	set.Add("s1", 1)
	set.Add("s2", 2)
	set.Add("s3", 3)
	set.Add("s4", 4)

	elems := set.PopMin(2)

	if elems[0].Member != "s1" || elems[1].Member != "s2" {
		t.Fail()
	}
}
func TestSortedSet_PopMax(t *testing.T) {
	var set = Make()
	set.Add("s1", 1)
	set.Add("s2", 2)
	set.Add("s3", 3)
	set.Add("s4", 4)

	elems := set.PopMax(2)
	if elems[0].Member != "s3" || elems[1].Member != "s4" {
		t.Fail()
	}

	set = Make()
	set.Add("s1", 1)
	set.Add("s2", 2)
	set.Add("s3", 3)
	set.Add("s4", 4)
	elems = set.PopMax(5)
	if len(elems) != 4 {
		t.Errorf("length should be 4 but now is %d", len(elems))
	}
}

func TestSortedSet_Count(t *testing.T) {
	var set = Make()
	set.Add("s1", 1)
	set.Add("s2", 2)
	set.Add("s3", 3)
	set.Add("s4", 4)

	var minBorder = &ScoreBorder{
		Value:   -1,
		Exclude: false,
	}

	var maxBorder = &ScoreBorder{
		Value:   3,
		Exclude: true,
	}
	// [-1,3) result should be 2
	var count = set.Count(minBorder, maxBorder)
	if count != 2 {
		t.Fail()
	}
}
