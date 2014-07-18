package main

import (
	"fmt"
)

type Stepper struct {
	iperm   Index
	index   Index
	operm   Index
	pattern Triple
	pred    func(Triple) bool
	emit    func([]Triple) interface{}
}

func (g *Graph) Walk(at Triple, ss []Stepper) *chan interface{} {
	c := make(chan interface{})
	go g.Launch(&c, at, ss)
	return &c
}

func (g *Graph) Launch(c *chan interface{}, at Triple, ss []Stepper) {
	g.Step(c, []Triple{at}, ss)
	*c <- nil
}

func last(ts []Triple) *Triple {
	t := ts[len(ts)-1]
	return &t
}

func (g *Graph) Step(c *chan interface{}, ts []Triple, ss []Stepper) {
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

func Out(p []byte) Stepper {
	return Stepper{OPS, SPO, SPO, Triple{nil, p, nil, nil}, nil, nil}
}

func In(p []byte) Stepper {
	return Stepper{OPS, OPS, OPS, Triple{nil, p, nil, nil}, nil, nil}
}

func Has(pred func(Triple) bool) Stepper {
	return Stepper{SPO, SPO, SPO, Triple{}, pred, nil}
}

func (s Stepper) Emitter(f func([]Triple) interface{}) Stepper {
	return Stepper{s.iperm, s.index, s.operm, s.pattern, s.pred, f}
}

func PathToString(ts []Triple) string {
	acc := "Path:"
	for _, t := range ts {
		acc += fmt.Sprintf(" %v", t.ToStrings())
	}
	return acc
}
