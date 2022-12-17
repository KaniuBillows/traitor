package list

type Expected func(a any) bool

type Consumer func(idx int, val any) bool

type List interface {
	Add(val any)
	Get(index int) (val any)
	Set(index int, val any)
	Insert(index int, val any)
	Remove(index int) (val any)
	RemoveLast() (val any)
	RemoveAllByVal(expected Expected) int
	ReverseRemoveByVal(expected Expected, count int) int
	RemoveByVal(expected Expected, count int) int
	Len() int
	ForEach(consumer Consumer)
	Contains(expected Expected) bool
	Range(start int, stop int) []any
}
