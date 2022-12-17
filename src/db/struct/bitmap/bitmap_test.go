package bitmap

import (
	"bytes"
	"fmt"
	"io"
	"testing"
)

type s1 struct {
	id string
}
type s2 struct {
	s1
	name string
}

func (s1 *s1) f1() {

}
func (s2 *s2) f1() {
	var s = &bytes.Buffer{}
	var i io.Reader
	i = s
	fmt.Println(i)
}
func TestName(t *testing.T) {
	s := s2{
		s1:   s1{id: "123"},
		name: "123",
	}
	s.f1()
	ss1 := s1{
		id: "123",
	}
	ss1.f1()
}
