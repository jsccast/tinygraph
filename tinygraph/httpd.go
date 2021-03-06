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

package main

// Simple HTTP server that handles Javascript.  A single Javascript
// interpreter is shared, which is not ideal.

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/robertkrimen/otto"
	. "github.csv.comcast.com/jsteph206/tinygraph"
)

// We have a single Javascript interpreter, which we probably shouldn't.
var httpVM *otto.Otto

func runHttpd() {
	log.Printf("Opening config %s", *configFile)
	SharedGraph, _ = GetGraph(*configFile)
	http.HandleFunc("/js", handleJavascript)
	log.Printf("Start HTTP server %s", *httpPort)
	log.Printf("Done with HTTP server (%v)", http.ListenAndServe(*httpPort, nil))
}

func handleJavascript(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	js := r.FormValue("js")
	if js == "" {
		js = "{};"
	}
	pretty := r.FormValue("pretty")
	switch pretty {
	case "":
		pretty = "true"
	case "true":
	default:
		pretty = "false"
	}

	log.Printf("javascript: executing %s\n", js)

	var vm *otto.Otto
	if *sharedHttpVM {
		if httpVM == nil {
			httpVM = otto.New()
			InitEnv(httpVM)
		}
		vm = httpVM
	} else {
		vm = otto.New()
		InitEnv(vm)
	}

	o, err := vm.Run(js)

	if err != nil {
		log.Printf("javascript: warning: user error: %v", err)
		fmt.Fprintf(w, `{"error":"%v", "js":"%s"}`, err, js)
		return
	}

	x, err := o.Export()

	// fmt.Fprintf(os.Stderr, "got %v (%v)\n", x, err)
	if err != nil {
		log.Printf("javascript: warning: export error: %v", err)
		fmt.Fprintf(w, `{"error":"%v", "js":"%s"}`, err, js)
		return
	}
	var bs []byte
	if pretty == "true" {
		bs, err = json.MarshalIndent(&x, "  ", "  ")
	} else {
		bs, err = json.Marshal(&x)
	}
	if err != nil {
		log.Printf("javascript: warning: marshal error: %v", err)
		fmt.Fprintf(w, `{"error":"%v", "js":"%s"}`, err, js)
		return
	}

	log.Printf("javascript: returning %d bytes\n", len(bs))

	fmt.Fprintf(w, "%s\n", bs)
}
