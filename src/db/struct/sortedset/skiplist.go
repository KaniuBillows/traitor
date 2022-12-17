package sortedset

import (
	"math/bits"
	"math/rand"
)

const (
	maxLevel = 16
)

type Element struct {
	Member string
	Score  float64
}

type node struct {
	Element
	backward *node
	level    []*Level
}
type Level struct {
	forward *node
	span    int64
}

type skipList struct {
	header *node
	tail   *node
	length int64
	level  int16
}

func makeNode(level int16, score float64, member string) *node {
	n := &node{
		Element: Element{
			Score:  score,
			Member: member,
		},
		level: make([]*Level, level),
	}
	for i := range n.level {
		n.level[i] = new(Level)
	}
	return n
}
func makeSkipList() *skipList {
	return &skipList{
		level:  1,
		header: makeNode(maxLevel, 0, ""),
	}
}
func randomLevel() int16 {
	total := uint64(1)<<uint64(maxLevel) - 1
	k := rand.Uint64() % total
	return maxLevel - int16(bits.Len64(k)) + 1
}

func (s *skipList) insert(member string, score float64) *node {
	update := make([]*node, maxLevel) // link new node with node in `update`
	rank := make([]int64, maxLevel)

	// find position to insert
	node := s.header
	for i := s.level - 1; i >= 0; i-- {
		if i == s.level-1 {
			rank[i] = 0
		} else {
			rank[i] = rank[i+1] // store rank that is crossed to reach the insert position
		}
		if node.level[i] != nil {
			// traverse the skip list
			for node.level[i].forward != nil &&
				(node.level[i].forward.Score < score ||
					(node.level[i].forward.Score == score && node.level[i].forward.Member < member)) { // same score, different key
				rank[i] += node.level[i].span
				node = node.level[i].forward
			}
		}
		update[i] = node
	}

	level := randomLevel()
	// extend skiplist level
	if level > s.level {
		for i := s.level; i < level; i++ {
			rank[i] = 0
			update[i] = s.header
			update[i].level[i].span = s.length
		}
		s.level = level
	}

	// make node and link into skiplist
	node = makeNode(level, score, member)
	for i := int16(0); i < level; i++ {
		node.level[i].forward = update[i].level[i].forward
		update[i].level[i].forward = node

		// update span covered by update[i] as node is inserted here
		node.level[i].span = update[i].level[i].span - (rank[0] - rank[i])
		update[i].level[i].span = (rank[0] - rank[i]) + 1
	}

	// increment span for untouched levels
	for i := level; i < s.level; i++ {
		update[i].level[i].span++
	}

	// set backward node
	if update[0] == s.header {
		node.backward = nil
	} else {
		node.backward = update[0]
	}
	if node.level[0].forward != nil {
		node.level[0].forward.backward = node
	} else {
		s.tail = node
	}
	s.length++
	return node
}

/*
 * param node: node to delete
 * param update: backward node (of target)
 */
func (s *skipList) removeNode(node *node, update []*node) {
	for i := int16(0); i < s.level; i++ {
		if update[i].level[i].forward == node {
			update[i].level[i].span += node.level[i].span - 1
			update[i].level[i].forward = node.level[i].forward
		} else {
			update[i].level[i].span--
		}
	}
	if node.level[0].forward != nil {
		node.level[0].forward.backward = node.backward
	} else {
		s.tail = node.backward
	}
	for s.level > 1 && s.header.level[s.level-1].forward == nil {
		s.level--
	}
	s.length--
}

/*
 * return: has found and removed node
 */
func (s *skipList) remove(member string, score float64) bool {
	/*
	 * find backward node (of target) or last node of each level
	 * their forward need to be updated
	 */
	update := make([]*node, maxLevel)
	node := s.header
	for i := s.level - 1; i >= 0; i-- {
		for node.level[i].forward != nil &&
			(node.level[i].forward.Score < score ||
				(node.level[i].forward.Score == score &&
					node.level[i].forward.Member < member)) {
			node = node.level[i].forward
		}
		update[i] = node
	}
	node = node.level[0].forward
	if node != nil && score == node.Score && node.Member == member {
		s.removeNode(node, update)
		// free x
		return true
	}
	return false
}

// getRank
//
// return: 1 based rank
func (s *skipList) getRank(member string, score float64) (exists bool, rank int64) {
	rank = 0
	x := s.header
	for i := s.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil &&
			(x.level[i].forward.Score < score ||
				(x.level[i].forward.Score == score &&
					x.level[i].forward.Member <= member)) {
			rank += x.level[i].span
			x = x.level[i].forward
		}

		/* x might be equal to zsl->header, so test if obj is non-NULL */
		if x.Member == member {
			return true, rank
		}
	}
	return false, rank
}

/*
RemoveRangeByScore

return removed elements
*/
func (s *skipList) RemoveRangeByScore(min *ScoreBorder, max *ScoreBorder, limit int) (removed []*Element) {
	update := make([]*node, maxLevel)
	removed = make([]*Element, 0)
	// find backward nodes (of target range) or last node of each level
	node := s.header
	for i := s.level - 1; i >= 0; i-- {
		for node.level[i].forward != nil {
			if min.less(node.level[i].forward.Score) { // already in range
				break
			}
			node = node.level[i].forward
		}
		update[i] = node
	}

	// node is the first one within range
	node = node.level[0].forward

	// remove nodes in range
	for node != nil {
		if !max.greater(node.Score) { // already out of range
			break
		}
		next := node.level[0].forward
		removedElement := node.Element
		removed = append(removed, &removedElement)
		s.removeNode(node, update)
		if limit > 0 && len(removed) == limit {
			break
		}
		node = next
	}
	return removed
}

// RemoveRangeByRank
//
// 1-based rank, including start, exclude stop
func (s *skipList) RemoveRangeByRank(start int64, stop int64) (removed []*Element) {
	var i int64 = 0 // rank of iterator
	update := make([]*node, maxLevel)
	removed = make([]*Element, 0)

	// scan from top level
	node := s.header
	for level := s.level - 1; level >= 0; level-- {
		for node.level[level].forward != nil && (i+node.level[level].span) < start {
			i += node.level[level].span
			node = node.level[level].forward
		}
		update[level] = node
	}

	i++
	node = node.level[0].forward // first node in range

	// remove nodes in range
	for node != nil && i < stop {
		next := node.level[0].forward
		removedElement := node.Element
		removed = append(removed, &removedElement)
		s.removeNode(node, update)
		node = next
		i++
	}
	return removed
}
func (s *skipList) getByRank(rank int64) *node {
	var i int64 = 0
	n := s.header
	for level := s.level - 1; level >= 0; level-- {
		// searching in the current level.
		// if next node's rank is larger than rank, go to next level.
		for n.level[level].forward != nil && // next node's rank means: current i plus next node's span.
			(i+n.level[level].span) <= rank {
			// it means the rank one is in the current level.
			// we just move to next one by one.
			i += n.level[level].span   // update current rank.
			n = n.level[level].forward //just like normal linked list. move to next.
		}
		if i == rank { // current rank match target.
			return n
		}
	}
	// all level search finished. but not match target.
	return nil
}
func (s *skipList) hasInRange(min *ScoreBorder, max *ScoreBorder) bool {
	// min & max = empty
	if min.Value > max.Value || (min.Value == max.Value && (min.Exclude || max.Exclude)) {
		return false
	}
	if s.header.level[0].forward == nil || s.tail == nil {
		return false
	}
	var mi = min.less(s.tail.Score)
	var mx = max.greater(s.header.level[0].forward.Score)
	return mi && mx
}

// list: 0 1 2 3
// border: [1,3)
// result 1
func (s *skipList) getFirstInScoreRange(min *ScoreBorder, max *ScoreBorder) *node {
	if s.hasInRange(min, max) == false {
		return nil
	}
	n := s.header
	for level := s.level - 1; level >= 0; level-- {
		for n.level[level].forward != nil && !min.less(n.level[level].forward.Score) {
			n = n.level[level].forward
		}
	}
	n = n.level[0].forward
	if !max.greater(n.Score) {
		return nil
	}
	return n
}

// list: 0 1 2 3
// border: [1,3)
// result 2
func (s *skipList) getLastInScoreRange(min *ScoreBorder, max *ScoreBorder) *node {
	if s.hasInRange(min, max) == false {
		return nil
	}
	n := s.header
	for level := s.level - 1; level >= 0; level-- {
		for n.level[level].forward != nil && max.greater(n.level[level].forward.Score) {
			n = n.level[level].forward
		}
	}
	if !min.less(n.Score) {
		return nil
	}
	return n
}
