package list

type LinkedList struct {
	first *node
	last  *node
	size  int
}

type node struct {
	val  any
	prev *node
	next *node
}

func Make(values ...any) *LinkedList {
	list := &LinkedList{
		first: nil,
		last:  nil,
		size:  0,
	}
	for _, v := range values {
		list.Add(v)
	}
	return list
}
func (l *LinkedList) Add(val any) {
	n := &node{
		val: val,
	}
	if l.last == nil {
		l.first = n
		l.last = n
	} else {
		l.last.next = n
		n.prev = l.last
		l.last = n
	}
	l.size++
}

func (l *LinkedList) find(index int) (n *node) {
	if index < 0 || index >= l.size {
		panic("index out of bound")
	}
	if index < l.size/2 {
		n = l.first
		for i := 0; i < index; i++ {
			n = n.next
		}
	} else {
		n = l.last
		for i := l.size - 1; i > index; i-- {
			n = n.prev
		}
	}
	return n
}

func (l *LinkedList) Get(index int) (val any) {
	return l.find(index).val
}

func (l *LinkedList) Set(index int, val any) {
	n := l.find(index)
	n.val = val
}

func (l *LinkedList) Insert(index int, val any) {
	if index == l.size {
		l.Add(val)
	}

	p := l.find(index)
	n := &node{
		val:  val,
		prev: p.prev,
		next: p,
	}
	if p.prev == nil {
		l.first = n
	} else {
		p.prev.next = n
	}
	p.prev = n
	l.size++
}

func (l *LinkedList) removeNode(n *node) {
	if n.prev == nil {
		l.first = n.next
	} else {
		n.prev.next = n.next
	}
	if n.next == nil {
		l.last = n.prev
	} else {
		n.next.prev = n.prev
	}
	n.prev = nil
	n.next = nil

	l.size--
}

func (l *LinkedList) Remove(index int) (val any) {
	n := l.find(index)
	l.removeNode(n)
	return n.val
}

func (l *LinkedList) RemoveLast() (val any) {
	n := l.last
	if n == nil {
		return nil
	}
	l.removeNode(n)
	return n.val
}

func (l *LinkedList) RemoveAllByVal(expected Expected) (removed int) {
	n := l.first
	removed = 0
	var next *node
	for n != nil {
		next = n.next
		if expected(n.val) {
			l.removeNode(n)
			removed++
		}
		n = next
	}
	return removed
}

func (l *LinkedList) RemoveByVal(expected Expected, count int) (removed int) {
	n := l.first
	removed = 0
	var next *node
	for n != nil {
		next = n.next
		if expected(n.val) {
			l.removeNode(n)
			removed++
		}
		if removed == count {
			break
		}
		n = next
	}
	return removed
}
func (l *LinkedList) ReverseRemoveByVal(expected Expected, count int) int {
	if l == nil {
		panic("l is nil")
	}
	n := l.last
	removed := 0
	var prevNode *node
	for n != nil {
		prevNode = n.prev
		if expected(n.val) {
			l.removeNode(n)
			removed++
		}
		if removed == count {
			break
		}
		n = prevNode
	}
	return removed
}

func (l *LinkedList) Len() int {
	return l.size
}

func (l *LinkedList) ForEach(consumer Consumer) {
	n := l.first
	i := 0
	for n != nil {
		ctu := consumer(i, n.val)
		if ctu == false {
			break
		}
		i++
		n = n.next
	}
}

func (l *LinkedList) Contains(expected Expected) bool {
	res := false
	l.ForEach(func(idx int, val any) bool {
		if expected(val) {
			res = true
			return false
		}
		return true
	})
	return res
}

func (l *LinkedList) Range(start int, stop int) []any {
	if start < 0 || start >= l.size {
		panic("`start` out of range")
	}
	if stop < start || stop > l.size {
		panic("`stop` out of range")
	}

	size := stop - start
	slice := make([]any, size)

	n := l.first
	i := 0
	for n != nil {
		if i >= start && i < stop {
			slice[i] = n.val
		} else if i >= stop {
			break
		}
		i++
		n = n.next
	}
	return slice
}
