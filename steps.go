package main

import (
	"fmt"
)

type Stepper struct {
	iperm    Index
	index    Index
	operm    Index
	pattern  Triple
	pred     func(Triple) bool
	emit     func([]Triple) interface{}
	previous *Stepper
}

func (g *Graph) Walk(at Triple, ss []*Stepper) *chan interface{} {
	c := make(chan interface{})
	go g.Launch(&c, at, ss)
	return &c
}

func (g *Graph) Launch(c *chan interface{}, at Triple, ss []*Stepper) {
	g.Step(c, []Triple{at}, ss)
	*c <- nil
}

func last(ts []Triple) *Triple {
	t := ts[len(ts)-1]
	return &t
}

func (g *Graph) Step(c *chan interface{}, ts []Triple, ss []*Stepper) {
	at := last(ts)
	if len(ss) == 0 {
		// *c <- at
	} else {
		s := ss[0]
		if s.pred != nil {
			if s.pred(*at) {
				if s.emit != nil {
					*c <- s.emit(ts[1:])
				}
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
				if s.emit != nil {
					*c <- s.emit(path[1:])
				}
				g.Step(c, path, ss[1:])
			}
		}
	}
}

func Out(p []byte) *Stepper {
	return &Stepper{OPS, SPO, SPO, Triple{nil, p, nil, nil}, nil, nil, nil}
}

func (s *Stepper) Out(p []byte) *Stepper {
	next := Out(p)
	next.previous = s
	return next
}

func In(p []byte) *Stepper {
	return &Stepper{OPS, OPS, OPS, Triple{nil, p, nil, nil}, nil, nil, nil}
}

func Has(pred func(Triple) bool) *Stepper {
	return &Stepper{SPO, SPO, SPO, Triple{}, pred, nil, nil}
}

func (s *Stepper) Has(pred func(Triple) bool) *Stepper {
	next := Has(pred)
	next.previous = s
	return next
}

func (s *Stepper) In(p []byte) *Stepper {
	next := In(p)
	next.previous = s
	return next
}

func (s *Stepper) Emitter(f func([]Triple) interface{}) *Stepper {
	replacement := &Stepper{s.iperm, s.index, s.operm, s.pattern, s.pred, f, nil}
	replacement.previous = s.previous
	return replacement
}

func PathToString(ts []Triple) string {
	acc := "Path:"
	for _, t := range ts {
		acc += fmt.Sprintf(" %v", t.ToStrings())
	}
	return acc
}

func (s *Stepper) Walk(g *Graph, from Triple) *chan interface{} {
	at := s
	fmt.Printf("%p %v\n", at, *at)
	ss := make([]*Stepper, 0, 1)
	ss = append(ss, at)
	for at.previous != nil {
		ss = append([]*Stepper{at.previous}, ss...)
		at = at.previous
		fmt.Printf("%p %v\n", at, *at)
	}
	fmt.Printf("steps %v\n", ss)
	return g.Walk(from, ss)
}

func StepsTest(g *Graph) {
	has := func(t Triple) bool {
		return true
	}

	t := Triple{nil, nil, []byte("a"), nil}
	c := g.Walk(t, []*Stepper{
		Out([]byte("p1")).Emitter(func(ts []Triple) interface{} { return "p1:" + string(last(ts).O) }),
		Out([]byte("p2")),
		In([]byte("p4")),
		Out([]byte("p1")),
		Has(has).Emitter(func(ts []Triple) interface{} { return PathToString(ts) + " last" })})
	for {
		x := <-*c
		if x == nil {
			break
		}
		fmt.Printf("got %v\n", x)
	}

	fmt.Printf("again\n")

	t = Triple{nil, nil, []byte("a"), nil}
	c = Out([]byte("p1")).Emitter(func(ts []Triple) interface{} { return "p1:" + string(last(ts).O) }).
		Out([]byte("p2")).
		In([]byte("p4")).
		Out([]byte("p1")).Has(has).Emitter(func(ts []Triple) interface{} { return PathToString(ts) + " last" }).
		Walk(g, t)

	for {
		x := <-*c
		if x == nil {
			break
		}
		fmt.Printf("got %v\n", x)
	}

}
