package main

import (
	"bufio"
	"fmt"
	"github.com/robertkrimen/otto"
	"os"
	"strings"
)

// g = G.Open('config.json')
// G.Out("p1").Walk(g, G.Vertex("a")).Collect()[0][0].ToStrings()

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

func (e *Env) Bs(s string) []byte {
	return []byte(s)
}

func (e *Env) Scan(g *Graph, s []byte, limit int64) [][]string {
	alloc := limit
	if 10000 < alloc {
		alloc = 10000
	}
	acc := make([][]string, 0, alloc)
	g.Do(SPO, &Triple{[]byte(s), nil, nil, nil}, nil,
		func(t *Triple) bool {
			acc = append(acc, t.ToStrings())
			limit--
			if limit == 0 {
				return false
			}
			return true
		})
	return acc
}

func initEnv(vm *otto.Otto) {
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
	initEnv(vm)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "quit()" {
			break
		}
		x, err := vm.Run(line)
		fmt.Printf("%v (%v)\n", x, err)
	}
}
