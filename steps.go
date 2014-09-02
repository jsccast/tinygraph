package main

import (
	"fmt"
	"sync/atomic"
)

type Stepper struct {
	iperm    Index
	index    Index
	operm    Index
	pattern  Triple
	pred     func(Triple) bool
	fs       []func(Path)
	previous *Stepper
}

type Chants chan []Triple

// We wrap because Otto wants us to.
type Chan struct {
	c     Chants
	state uint32
}

func NewChan() *Chan {
	return &Chan{make(Chants), Open}
}

const (
	Open uint32 = iota

	// ToDo: No this.  Need another channel for control.
	Closed
)

func (c *Chan) IsClosed() bool {
	return Closed == atomic.LoadUint32(&c.state)
}

func (c *Chan) Close() {
	fmt.Printf("closing chan %p\n", c)
	atomic.StoreUint32(&c.state, Closed)
	close(c.c)
}

type Vertex []byte

type Path []Triple

func (v *Vertex) ToTriple() Triple {
	return Triple{nil, nil, *v, nil}
}

func (g *Graph) Walk(o Vertex, ss []*Stepper) *Chan {
	c := NewChan()
	go g.Launch(c, o.ToTriple(), ss)
	return c
}

func (g *Graph) Launch(c *Chan, at Triple, ss []*Stepper) {
	g.Step(c, Path{at}, ss)
	if !c.IsClosed() {
		(*c).c <- nil
	}
}

func last(ts Path) *Triple {
	t := ts[len(ts)-1]
	return &t
}

func (s *Stepper) exec(path Path) {
	for _, f := range s.fs {
		f(path)
	}
}

func (g *Graph) Step(c *Chan, ts Path, ss []*Stepper) bool {
	if c.IsClosed() {
		fmt.Printf("chan %p is closed\n", c)
		return false
	}
	if len(ss) == 0 {
		fmt.Printf("chan %p closed? %v\n", c.IsClosed())
		(*c).c <- ts[1:]
	} else {
		at := last(ts)
		s := ss[0]
		if s.pred != nil {
			if s.pred(*at) {
				s.exec(ts[1:])
				g.Step(c, ts, ss[1:])
			}
		} else {
			u := at.Permute(s.iperm)
			if s.pattern.S != nil {
				u.S = s.pattern.S
			}
			if s.pattern.P != nil {
				u.P = s.pattern.P
			}
			u.O = nil
			i := g.NewIndexIterator(s.index, u, nil)
			for i.Next() {
				t := IndexedTripleFromBytes(s.index, i.Key(), i.Value())
				t = t.Permute(s.index).Permute(s.operm)
				path := append(ts, *t)
				s.exec(path[1:])
				if !g.Step(c, path, ss[1:]) {
					return false
				}
			}
		}
	}

	return true
}

func Out(p []byte) *Stepper {
	return &Stepper{OPS, SPO, SPO, Triple{nil, p, nil, nil}, nil, make([]func(Path), 0, 0), nil}
}

func (s *Stepper) Out(p []byte) *Stepper {
	next := Out(p)
	next.previous = s
	return next
}

func AllOut() *Stepper {
	return &Stepper{OPS, SPO, SPO, Triple{nil, nil, nil, nil}, nil, make([]func(Path), 0, 0), nil}
}

func (s *Stepper) AllOut() *Stepper {
	next := AllOut()
	next.previous = s
	return next
}

func In(p []byte) *Stepper {
	return &Stepper{OPS, OPS, OPS, Triple{nil, p, nil, nil}, nil, make([]func(Path), 0, 0), nil}
}

func (s *Stepper) In(p []byte) *Stepper {
	next := In(p)
	next.previous = s
	return next
}

func AllIn() *Stepper {
	return &Stepper{OPS, OPS, OPS, Triple{nil, nil, nil, nil}, nil, make([]func(Path), 0, 0), nil}
}

func (s *Stepper) AllIn() *Stepper {
	next := AllIn()
	next.previous = s
	return next
}

func Has(pred func(Triple) bool) *Stepper {
	return &Stepper{SPO, SPO, SPO, Triple{}, pred, make([]func(Path), 0, 0), nil}
}

func (s *Stepper) Has(pred func(Triple) bool) *Stepper {
	next := Has(pred)
	next.previous = s
	return next
}

// func (s *Stepper) Emitter(f func(Path) interface{}) *Stepper {
// 	replacement := &Stepper{s.iperm, s.index, s.operm, s.pattern, s.pred, f, nil}
// 	replacement.previous = s.previous
// 	return replacement
// }

// func (s *Stepper) Emit() *Stepper {
// 	return s.Emitter(func(ts Path) interface{} { return ts })
// }

func (s *Stepper) Do(f func(Path)) *Stepper {
	s.fs = append(s.fs, f)
	return s
}

func (path *Path) ToString() string {
	acc := "Path:"
	for _, t := range *path {
		acc += fmt.Sprintf(" %v", t.ToString())
	}
	return acc
}

func (s *Stepper) Walk(g *Graph, from Vertex) *Chan {
	at := s
	ss := make([]*Stepper, 0, 1)
	ss = append(ss, at)
	for at.previous != nil {
		ss = append([]*Stepper{at.previous}, ss...)
		at = at.previous
	}
	return g.Walk(from, ss)
}

func (c *Chan) Do(f func(Path)) {
	for {
		x := <-(*c).c
		if x == nil {
			break
		}
		f(x)
	}
	c.Close()
}

func (c *Chan) DoSome(f func(Path), limit int64) {
	n := int64(0)
	for {
		x := <-(*c).c
		if x == nil || n == limit {
			break
		}
		f(x)
		n++
	}
	c.Close()
}

func (c *Chan) Print() {
	c.Do(func(path Path) { fmt.Printf("path %v\n", path.ToString()) })
}

func (c *Chan) Collect() []Path {
	acc := make([]Path, 0, 0)
	c.Do(func(ts Path) {
		acc = append(acc, ts)
	})
	return acc
}

func (c *Chan) CollectSome(limit int64) []Path {
	acc := make([]Path, 0, 0)
	c.DoSome(func(ts Path) {
		acc = append(acc, ts)
	}, limit)
	return acc
}

func StepsTest() {

	g, _ := GetGraph("config.test")

	g.WriteIndexedTriple(TripleFromStrings("a", "p1", "b", "today"), nil)
	g.WriteIndexedTriple(TripleFromStrings("a", "p1", "f", "today"), nil)
	g.WriteIndexedTriple(TripleFromStrings("a", "p5", "j", "today"), nil)
	g.WriteIndexedTriple(TripleFromStrings("b", "p2", "c", "today"), nil)
	g.WriteIndexedTriple(TripleFromStrings("c", "p3", "d", "today"), nil)
	g.WriteIndexedTriple(TripleFromStrings("c", "p3", "e", "today"), nil)
	g.WriteIndexedTriple(TripleFromStrings("g", "p4", "c", "today"), nil)
	g.WriteIndexedTriple(TripleFromStrings("g", "p1", "h", "today"), nil)
	g.WriteIndexedTriple(TripleFromStrings("g", "p1", "i", "today"), nil)

	has := func(t Triple) bool {
		return string(t.O) == "i"
	}

	v := Vertex([]byte("a"))
	g.Walk(v, []*Stepper{AllOut()}).Print()

	v = Vertex([]byte("a"))
	g.Walk(v, []*Stepper{
		Out([]byte("p1")),
		Out([]byte("p2")),
		In([]byte("p4")),
		Out([]byte("p1")),
		Has(has)}).Print()

	fmt.Printf("again\n")

	Out([]byte("p1")).
		Out([]byte("p2")).
		In([]byte("p4")).
		Out([]byte("p1")).
		Has(has).
		Walk(g, v).
		Print()

	fmt.Printf("again again\n")

	Out(nil).
		Out([]byte("p2")).
		In([]byte("p4")).
		Out([]byte("p1")).
		Has(has).
		Walk(g, v).
		Print()

	fmt.Printf("map print\n")

	Out(nil).
		Do(func(path Path) { fmt.Printf("doing s %v\n", path.ToString()) }).
		Out([]byte("p2")).
		Do(func(path Path) { fmt.Printf("doing 0 %v\n", path.ToString()) }).
		In([]byte("p4")).
		Do(func(path Path) { fmt.Printf("doing 1 %v\n", path.ToString()) }).
		Out([]byte("p1")).
		Has(func(t Triple) bool { return string(t.O) == "i" }).
		Walk(g, v).
		Do(func(path Path) { fmt.Printf("got %v\n", path.ToString()) })

	fmt.Printf("for repl print\n")

	Out([]byte("p1")).
		Walk(g, v).
		Print()

	fmt.Printf("collected %v\n", Out([]byte("p1")).
		Walk(g, v).
		Collect())
}
