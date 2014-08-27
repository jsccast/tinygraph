package main

import (
	"encoding/json"
	"fmt"
	"github.com/robertkrimen/otto"
	"net/http"
	"os"
)

var httpdGraph *Graph
var httpVM *otto.Otto

func runHttpd() {
	fmt.Fprintf(os.Stderr, "opening %s\n", *configFile)
	httpdGraph, _ = GetGraph(*configFile)
	http.HandleFunc("/js", handleJavascript)
	port := ":8080"
	fmt.Fprintf(os.Stderr, "start %s\n", port)
	fmt.Fprintf(os.Stderr, "done %v\n", http.ListenAndServe(port, nil))
}

// Sorry
func (e *Env) Graph() *Graph {
	return httpdGraph
}

func handleJavascript(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	js := r.FormValue("js")
	if js == "" {
		js = "{};"
	}
	fmt.Fprintf(os.Stderr, "executing %s\n", js)

	var vm *otto.Otto
	if *sharedHttpVM {
		if httpVM == nil {
			httpVM = otto.New()
			initEnv(httpVM)
		}
		vm = httpVM
	} else {
		vm = otto.New()
		initEnv(vm)
	}

	o, err := vm.Run(js)

	if err != nil {
		fmt.Fprintf(w, `{"error":"%v", "js":"%s"}`, err, js)
		return
	}

	x, err := o.Export()

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
