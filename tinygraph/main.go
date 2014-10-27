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

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	. "github.csv.comcast.com/jsteph206/tinygraph"
)

var filesToLoad = flag.String("load", "", "Files to load")
var repl = flag.Bool("repl", false, "Run REPL")
var serve = flag.Bool("serve", false, "Start HTTPD server")
var configFile = flag.String("config", "config.js", "Configuration file")
var sharedHttpVM = flag.Bool("sharevm", true, "Use a shared Javascript VM for the HTTP service")
var httpPort = flag.String("port", ":8080", "HTTP server port")

func RationalizeMaxProcs() {
	if os.Getenv("GOMAXPROCS") == "" {
		n := runtime.NumCPU()
		log.Printf("Setting GOMAXPROCS to %d\n", n)
		runtime.GOMAXPROCS(n)
	} else {
		log.Printf("GOMAXPROCS is %v\n", os.Getenv("GOMAXPROCS"))
	}
}

func WriteStatsLoop(g *Graph) {
	go func() {
		for {
			time.Sleep(10 * time.Second)
			bs := []byte(g.GetStats())
			ioutil.WriteFile("stats.log", bs, 0644)
		}
	}()
}

func Load() {
	g, config := GetGraph(*configFile)
	log.Println(g.GetStats())

	if b, ok := config.BoolKey("initial_compaction"); ok && b {
		g.Compact()
		log.Println(g.GetStats())
	}

	if b, ok := config.BoolKey("stats_loop"); ok && b {
		WriteStatsLoop(g)
	}

	wait := sync.WaitGroup{}
	for _, filename := range strings.Split(*filesToLoad, ",") {
		filename = strings.TrimSpace(filename)
		log.Printf("loading triples: %s\n", filename)
		wait.Add(1)
		go g.LoadTriplesFile(filename, config, &wait)
		// Stagger the threads a little.
		time.Sleep(1 * time.Second)
	}
	wait.Wait()

	log.Println(g.GetStats())

	err := g.Close()
	if err != nil {
		panic(err)
	}
}

func main() {
	flag.Parse()
	RationalizeMaxProcs()
	if *filesToLoad != "" {
		Load()
	}
	var wg sync.WaitGroup

	if *serve {
		wg.Add(1)
		go func() {
			runHttpd()
			wg.Done()
		}()
	}
	if *repl {
		wg.Add(1)
		go func() {
			REPL()
			wg.Done()
		}()
	}
	wg.Wait()
}

func DoPrint(g *Graph, index Index, label string, s string) bool {
	limit := 100
	found := 0
	g.Do(index, &Triple{[]byte(s), nil, nil, nil}, nil,
		func(t *Triple) bool {
			fmt.Printf("%s %v\n", label, t.Strings())
			found++
			limit--
			if limit == 0 {
				return false
			}
			return true
		})
	return 0 < found
}
