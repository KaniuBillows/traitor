package set

import "traitor/db/struct/dict"

type Set struct {
	dict dict.Dict
}

func Make(members ...string) (set *Set) {
	set = &Set{
		dict: dict.MakeSimple(),
	}
	return
}

func (s *Set) Add(val string) int {
	return s.dict.Put(val, nil)
}
func (s *Set) Remove(val string) int {
	return s.dict.Remove(val)
}

func (s *Set) Has(val string) (exists bool) {
	_, exists = s.dict.Get(val)
	return
}

func (s *Set) Len() int {
	return s.dict.Len()
}

func (s *Set) ToSlice() []string {
	return s.dict.Keys()
}

func (s *Set) ForEach(consumer func(member string) bool) {
	s.dict.ForEach(func(key string, val any) bool {
		return consumer(key)
	})
}

func (s *Set) Intersect(s2 *Set) (set *Set) {
	set = Make()
	// choose smaller one as loop.
	var smSet *Set
	var bgrSet *Set
	if s.Len() < s2.Len() {
		smSet = s
		bgrSet = s2
	} else {
		smSet = s2
		bgrSet = s
	}
	smSet.ForEach(func(m string) bool {
		if bgrSet.Has(m) {
			set.Add(m)
		}
		return true
	})
	return
}

func (s *Set) Union(s2 *Set) (set *Set) {
	set = Make()
	s.ForEach(func(member string) bool {
		set.Add(member)
		return true
	})
	s2.ForEach(func(member string) bool {
		set.Add(member)
		return true
	})
	return
}

func (s *Set) Diff(s2 *Set) (set *Set) {
	set = Make()
	s.ForEach(func(member string) bool {
		if s2.Has(member) == false {
			set.Add(member)
		}
		return true
	})
	return set
}

func (s *Set) RandomMembers(limit int) []string {
	return s.dict.RandomKeys(limit)
}
func (s *Set) RandomDistinctMembers(limit int) []string {
	return s.dict.RandomDistinctKeys(limit)
}
