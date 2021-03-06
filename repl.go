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

// Expose some Go functions to Javascript.

import (
	"bufio"
	"fmt"
	"github.com/robertkrimen/otto"
	"os"
	"strings"
)

// g = G.Open('config.json')
// G.Out("p1").Walk(g, G.Vertex("a")).Collect()[0][0].ToStrings()

// Env hold our bindings.
type Env struct {
}

func (e *Env) Vertex(s string) Vertex {
	return []byte(s)
}

func (e *Env) Triple(s, p, o, v string) *Triple {
	return &Triple{[]byte(s), []byte(p), []byte(o), []byte(v)}
}

func (e *Env) Open(config string) *Graph {
	g, opts := GetGraph(config)
	if opts == nil {
		// ToDo
		fmt.Printf("Open %s mystery error", config)
	}
	return g
}

func GetGraph(configFilename string) (*Graph, *Options) {
	config, err := LoadOptions(configFilename)
	if err != nil {
		panic(err)
	}

	opts := RocksOpts(config)
	opts.SetCreateIfMissing(true)
	opts.SetErrorIfExists(false)

	dirname := "tmp.db"
	if dir, ok := config.StringKey("db_dir"); ok {
		dirname = dir
	}

	g, err := NewGraph(dirname, opts)

	if err != nil {
		panic(err)
	}

	g.wopts = RocksWriteOpts(config)
	g.ropts = RocksReadOpts(config)

	return g, config
}

var SharedGraph *Graph

// Graph returns the global graph.  Sorry.
func (e *Env) Graph() *Graph {
	return SharedGraph
}

func (e *Env) Out(p []byte) *Stepper {
	return Out(p)
}

func (e *Env) AllOut() *Stepper {
	return AllOut()
}

func (e *Env) In(p []byte) *Stepper {
	return In(p)
}

func (e *Env) AllIn() *Stepper {
	return AllIn()
}

// Bs converts the given string to a byte array.
func (e *Env) Bs(s string) []byte {
	return []byte(s)
}

// REPLIterator exposes an path channel as an iterator in Javascript.
// For convenience.  Might not be that useful.
type REPLIterator struct {
	c     *Chan
	n     int64
	state uint32
}

func (i *REPLIterator) Close() {
	if i.state != Closed {
		i.state = Closed
		i.c.Close()
	}
}

func (i *REPLIterator) IsClosed() bool {
	return i.state == Closed
}

func (i *REPLIterator) Next() interface{} {
	if i.n <= 0 {
		i.Close()
		return nil
	}
	i.n--
	return <-(i.c).c
}

// BUG(?): Iterator can return an array with a nil first component
// even though isClosed() is false.  See examples/iter.js.

func (c *Chan) Iter(limit int64) *REPLIterator {
	return &REPLIterator{c, limit, Open}
}

// Scan returns up to 'limit' triples starting with the given vertex.
func (e *Env) Scan(g *Graph, s []byte, limit int64) [][]string {
	alloc := limit
	if 10000 < alloc {
		alloc = 10000
	}
	acc := make([][]string, 0, alloc)
	g.Do(SPO, &Triple{[]byte(s), nil, nil, nil}, nil,
		func(t *Triple) bool {
			acc = append(acc, t.Strings())
			limit--
			if limit == 0 {
				return false
			}
			return true
		})
	return acc
}

func InitEnv(vm *otto.Otto) {
	vm.Set("G", new(Env))

	vm.Set("toJS", func(call otto.FunctionCall) otto.Value {
		result, err := vm.ToValue(call.Argument(0))
		if err != nil {
			panic(err)
		}
		return result
	})
}

func REPL() {
	scanner := bufio.NewScanner(os.Stdin)
	vm := otto.New()
	InitEnv(vm)
	// Complete statement/expression must be on one line.
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "quit()" {
			break
		}
		x, err := vm.Run(line)
		fmt.Printf("%v (%v)\n", x, err)
	}
}
