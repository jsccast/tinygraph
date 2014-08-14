package main

import (
	"bufio"
	"fmt"
	"github.com/robertkrimen/otto"
	"os"
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
	if opts != nil {
		fmt.Printf("Open %s error: %v\n", config)
	}
	return g
}

func (e *Env) Out(p []byte) *Stepper {
	return Out(p)
}

func (e *Env) In(p []byte) *Stepper {
	return In(p)
}

func (e *Env) Bs(s string) []byte {
	return []byte(s)
}

func REPL() {
	scanner := bufio.NewScanner(os.Stdin)
	vm := otto.New()
	vm.Set("G", new(Env))

	for scanner.Scan() {
		line := scanner.Text()
		x, err := vm.Run(line)
		fmt.Printf("%v (%v)\n", x, err)
	}
}
