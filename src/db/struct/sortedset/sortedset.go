package sortedset

import (
	"strconv"
)

type SortedSet struct {
	dict     map[string]*Element
	skipList *skipList
}

func Make() *SortedSet {
	return &SortedSet{
		dict:     make(map[string]*Element),
		skipList: makeSkipList(),
	}
}

// Add
//
// if the key is a new one,return true else return false.
func (s *SortedSet) Add(member string, score float64) bool {
	elem, ok := s.dict[member]
	s.dict[member] = &Element{
		Member: member,
		Score:  score,
	}
	if ok {
		if score != elem.Score {
			s.skipList.remove(member, elem.Score)
			s.skipList.insert(member, score)
		}
		return false
	}
	s.skipList.insert(member, score)
	return true
}

// Len
//
// Get member nums.
func (s *SortedSet) Len() int64 {
	return s.skipList.length
}

func (s *SortedSet) Get(member string) (element *Element, ok bool) {
	element, ok = s.dict[member]

	return element, true

}

func (s *SortedSet) Remove(member string) (exists bool) {
	elem, ok := s.dict[member]
	if ok == false {
		return false
	}
	s.skipList.remove(elem.Member, elem.Score)
	delete(s.dict, member)
	return ok
}

// GetRank
//
// Get key's rank
// result rank is 0-based.
// order by ascending order.
func (s *SortedSet) GetRank(member string, desc bool) (rank int64) {
	elem, ok := s.dict[member]
	if ok == false {
		return -1
	}
	ok, rank = s.skipList.getRank(elem.Member, elem.Score)
	if ok == false {
		return -1
	}
	if desc {
		rank = s.skipList.length - rank
	} else {
		rank--
	}
	return
}

func (s *SortedSet) Foreach(start int64, stop int64, desc bool, consumer func(elem *Element) bool) {
	size := int64(s.Len())
	if start < 0 || start >= size {
		panic("illegal start " + strconv.FormatInt(start, 10))
	}
	if stop < start || stop > size {
		panic("illegal end " + strconv.FormatInt(stop, 10))
	}

	var node *node
	if desc {
		node = s.skipList.tail
		if start > 0 {
			node = s.skipList.getByRank(size - start)
		}
	} else {
		node = s.skipList.header.level[0].forward
		if start > 0 {
			node = s.skipList.getByRank(start + 1)
		}
	}
	sliceSize := int(stop - start)
	for i := 0; i < sliceSize; i++ {
		if !consumer(&node.Element) {
			break
		}
		if desc {
			node = node.backward
		} else {
			node = node.level[0].forward
		}
	}
}

func (s *SortedSet) Range(start int64, stop int64, desc bool) []*Element {
	size := int(stop - start)
	slice := make([]*Element, size)
	i := 0
	s.Foreach(start, stop, desc, func(element *Element) bool {
		slice[i] = element
		i++
		return true
	})
	return slice
}

func (s *SortedSet) Count(min *ScoreBorder, max *ScoreBorder) int64 {
	var i int64 = 0
	s.Foreach(0, s.skipList.length, false, func(elem *Element) bool {
		gtMin := min.less(elem.Score)
		if gtMin == false {
			return true
		}
		ltMax := max.greater(elem.Score)
		if ltMax == false {
			return false
		}
		i++
		return true
	})
	return i
}

// ForEachByScore visits members which score within the given border
func (s *SortedSet) ForEachByScore(min *ScoreBorder, max *ScoreBorder, offset int64, limit int64, desc bool, consumer func(element *Element) bool) {
	// find start node
	var node *node
	if desc {
		node = s.skipList.getLastInScoreRange(min, max)
	} else {
		node = s.skipList.getFirstInScoreRange(min, max)
	}

	for node != nil && offset > 0 {
		if desc {
			node = node.backward
		} else {
			node = node.level[0].forward
		}
		offset--
	}

	// A negative limit returns all elements from the offset
	for i := 0; (i < int(limit) || limit < 0) && node != nil; i++ {
		if !consumer(&node.Element) {
			break
		}
		if desc {
			node = node.backward
		} else {
			node = node.level[0].forward
		}
		if node == nil {
			break
		}
		gtMin := min.less(node.Element.Score) // greater than min
		ltMax := max.greater(node.Element.Score)
		if !gtMin || !ltMax {
			break // break through score border
		}
	}
}

func (s *SortedSet) RangeByScore(min *ScoreBorder, max *ScoreBorder, offset int64, limit int64, desc bool) []*Element {
	if limit == 0 || offset < 0 {
		return make([]*Element, 0)
	}
	slice := make([]*Element, 0)
	s.ForEachByScore(min, max, offset, limit, desc, func(element *Element) bool {
		slice = append(slice, element)
		return true
	})
	return slice
}

// RemoveByScore removes members which score within the given border
func (s *SortedSet) RemoveByScore(min *ScoreBorder, max *ScoreBorder) int64 {
	removed := s.skipList.RemoveRangeByScore(min, max, 0)
	for _, element := range removed {
		delete(s.dict, element.Member)
	}
	return int64(len(removed))
}
func (s *SortedSet) RemoveByRank(start int64, end int64) int64 {
	removed := s.skipList.RemoveRangeByRank(start+1, end+1)
	for _, elem := range removed {
		delete(s.dict, elem.Member)
	}
	return int64(len(removed))
}
func (s *SortedSet) PopMin(count int) []*Element {
	first := s.skipList.getFirstInScoreRange(negativeInfBorder, positiveInfBorder)
	if first == nil {
		return nil
	}
	border := &ScoreBorder{
		Value:   first.Score,
		Exclude: false,
	}
	removed := s.skipList.RemoveRangeByScore(border, positiveInfBorder, count)
	for _, element := range removed {
		delete(s.dict, element.Member)
	}
	return removed
}

func (s *SortedSet) PopMax(count int) []*Element {

	removed := s.skipList.RemoveRangeByRank(s.skipList.length-int64(count-1), s.skipList.length+1)

	for _, element := range removed {
		delete(s.dict, element.Member)
	}
	return removed
}
