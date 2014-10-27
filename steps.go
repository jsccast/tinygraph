// Copyright 2014 Comcast Cable Communications Management, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tinygraph

// This code provides a slightly higher-level interface to Graph.Do().
// In(), Out(), Do(), and Walk() are the top-level functions.  See
// 'steps_test.go' and 'examples/' for some examples.

// ToDo: Unexpose some functions.
// ToDo: Examples in steps_test.go

import (
	"fmt"
)

// A Stepper defines what to do when walking a graph from some path.
// All the logic is given by the 'step' function, which possibly
// extends the current path.  A Stepper is a pretty low-level thing.
type Stepper struct {
	iperm    Index
	index    Index
	operm    Index
	pattern  Triple
	pred     func(Triple) bool
	fs       []func(Path)
	previous *Stepper
}

type Path []Triple

type Chants chan Path

type Vertex []byte

// We wrap because Otto wants us to.
type Chan struct {
	c    Chants
	done chan (struct{})
}

func NewChan() *Chan {
	return &Chan{make(Chants, *chanBufferSize), make(chan (struct{}))}
}

const (
	Open uint32 = iota

	// ToDo: No this.  Need another channel for control.
	Closed
)

func (c *Chan) Close() {
	close(c.done)
}

func (v *Vertex) toTriple() Triple {
	return Triple{nil, nil, *v, nil}
}

// Start walking.
func (g *Graph) launch(c *Chan, at Triple, ss []*Stepper) {
	g.step(c, Path{at}, ss)
	select {
	case <-c.done:
	default:
		(*c).c <- nil
	}
}

// Walk starts a graph walk at the given vertex.  You get a channel of
// paths.  A path is just an array of triples.  The traversal is
// defined by the given array of Steppers (such as In()s and Outs()).
func (g *Graph) Walk(o Vertex, ss []*Stepper) *Chan {
	c := NewChan()
	go g.launch(c, o.toTriple(), ss)
	return c
}

// Perform all stepper function invocation (if any).
func (s *Stepper) exec(path Path) {
	for _, f := range s.fs {
		f(path)
	}
}

func (g *Graph) step(c *Chan, ts Path, ss []*Stepper) bool {
	if len(ss) == 0 {
		select {
		case _ = <-c.done:
			return false
		default:
			(*c).c <- ts[1:]
		}
	} else {
		at := &ts[len(ts)-1]
		s := ss[0]
		if s.pred != nil {
			if s.pred(*at) {
				s.exec(ts[1:])
				g.step(c, ts, ss[1:])
			}
		} else {
			u := at.Permute(s.iperm)
			if s.pattern.S != nil {
				u.S = s.pattern.S
			}
			u.P = s.pattern.P
			u.O = nil
			i := g.NewIndexIterator(s.index, u, nil)
			for i.Next() {
				t := IndexedTripleFromBytes(s.index, i.Key(), i.Value())
				t = t.Permute(s.index).Permute(s.operm)
				path := append(ts, *t)
				s.exec(path[1:])
				if !g.step(c, path, ss[1:]) {
					return false
				}
			}
		}
	}

	return true
}

// Out returns a Stepper that traverses all edges out of the Stepper's input verticies.
func Out(p []byte) *Stepper {
	return &Stepper{OPS, SPO, SPO, Triple{nil, p, nil, nil}, nil, make([]func(Path), 0, 0), nil}
}

// Out extends the stepper to follow out-bound edges with the given property.
func (s *Stepper) Out(p []byte) *Stepper {
	next := Out(p)
	next.previous = s
	return next
}

// AllOut returns a Stepper that traverses all out-bound edges.
func AllOut() *Stepper {
	return &Stepper{OPS, SPO, SPO, Triple{nil, nil, nil, nil}, nil, make([]func(Path), 0, 0), nil}
}

// AllOut extends the stepper to follow all edges.
func (s *Stepper) AllOut() *Stepper {
	next := AllOut()
	next.previous = s
	return next
}

// In returns a Stepper that traverses all edges into of the Stepper's input verticies.
func In(p []byte) *Stepper {
	return &Stepper{OPS, OPS, OPS, Triple{nil, p, nil, nil}, nil, make([]func(Path), 0, 0), nil}
}

// In extends the stepper to follow all in-bound edges with the given property.
func (s *Stepper) In(p []byte) *Stepper {
	next := In(p)
	next.previous = s
	return next
}

// AllIn returns a Stepper that traverses all in-bound edges.
func AllIn() *Stepper {
	return &Stepper{OPS, OPS, OPS, Triple{nil, nil, nil, nil}, nil, make([]func(Path), 0, 0), nil}
}

// AllIn extends the stepper to follow all in-bound edges.
func (s *Stepper) AllIn() *Stepper {
	next := AllIn()
	next.previous = s
	return next
}

// Has returns a stepper that will follow edges for which pred returns true.
func Has(pred func(Triple) bool) *Stepper {
	return &Stepper{SPO, SPO, SPO, Triple{}, pred, make([]func(Path), 0, 0), nil}
}

// Has extends a stepper to will follow edges for which pred returns true.
func (s *Stepper) Has(pred func(Triple) bool) *Stepper {
	next := Has(pred)
	next.previous = s
	return next
}

// Do extends a stepper to execute a the given function for the current path.
func (s *Stepper) Do(f func(Path)) *Stepper {
	s.fs = append(s.fs, f)
	return s
}

func (path *Path) String() string {
	acc := "["
	for _, t := range *path {
		acc += `"` + t.String() + `"`
	}
	acc += "]"
	return acc
}

// Walk starts the stepper at the given vertex.  Returns a channel of paths.
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

// Do is a utility function to call the given function on every path from the channel.
func (c *Chan) Do(f func(Path)) {
	defer c.Close()
	for {
		x := <-(*c).c
		if x == nil {
			break
		}
		f(x)
	}
}

// DoSome is a utility function to call the given function on paths
// from the channel at most 'limit' times.
func (c *Chan) DoSome(f func(Path), limit int64) {
	defer c.Close()
	n := int64(0)
	for {
		x := <-(*c).c
		if x == nil || n == limit {
			break
		}
		f(x)
		n++
	}
}

// Print is a utility function that prints every path from the channel.
func (c *Chan) Print() {
	c.Do(func(path Path) { fmt.Printf("path %s\n", path.String()) })
}

// Collect is a utility function that gathers up all of the paths into
// a single array.  Use caution.  See 'CollectSome()' and 'DoSome()'.
func (c *Chan) Collect() []Path {
	acc := make([]Path, 0, 0)
	c.Do(func(ts Path) {
		acc = append(acc, ts)
	})
	return acc
}

// CollectSome is a utility function that gathers up at most 'limit'
// paths into an array.
func (c *Chan) CollectSome(limit int64) []Path {
	acc := make([]Path, 0, 0)
	c.DoSome(func(ts Path) {
		acc = append(acc, ts)
	}, limit)
	return acc
}
