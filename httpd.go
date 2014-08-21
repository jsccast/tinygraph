package main

import (
	"encoding/json"
	"fmt"
	"github.com/robertkrimen/otto"
	"net/http"
	"os"
)

var httpdGraph *Graph

func init() {
	go func() {
		config := "httpd.js"
		if _, err := os.Stat(config); err == nil {
			fmt.Fprintf(os.Stderr, "opening %s\n", config)
			httpdGraph, _ = GetGraph(config)
			http.HandleFunc("/js", handleJavascript)
			port := ":8080"
			fmt.Fprintf(os.Stderr, "start %s\n", port)
			fmt.Fprintf(os.Stderr, "done %v\n", http.ListenAndServe(port, nil))
		}
	}()
}

// Sorry
func (e *Env) Graph() *Graph {
	return httpdGraph
}

func execute(js string) (interface{}, error) {
	vm := otto.New()
	initEnv(vm)

	o, err := vm.Run(js)
	if err != nil {
		return nil, err
	}
	return o.Export()
}

func handleJavascript(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	js := r.FormValue("js")
	if js == "" {
		js = "{};"
	}
	fmt.Fprintf(os.Stderr, "executing %s\n", js)
	x, err := execute(js)
	fmt.Fprintf(os.Stderr, "got %v (%v)\n", x, err)
	if err != nil {
		fmt.Fprintf(w, `{"error":"%v", "js":"%s"}`, err, js)
		return
	}
	bs, err := json.MarshalIndent(&x, "  ", "  ")
	if err != nil {
		fmt.Fprintf(w, `{"error":"%v", "js":"%s"}`, err, js)
		return
	}
	fmt.Fprintf(os.Stderr, "returning %s\n", bs)
	fmt.Fprintf(w, "%s\n", bs)
}
