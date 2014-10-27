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

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

func ReadTriplesFile(c chan *Triple, tripleFile string) error {
	f, err := os.Open(tripleFile)
	if err != nil {
		fmt.Printf("ReadTriplesFromFile: Couldn't open file %s: %v\n", tripleFile, err)
		close(c)
		return fmt.Errorf("Couldn't open file %s: %v", tripleFile, err)
	}

	defer func() {
		if err := f.Close(); err != nil {
			log.Printf("ReadTriplesFromFile Close error %v", err)
		}

	}()

	in := bufio.NewReader(f)

	if strings.HasSuffix(tripleFile, ".gz") || *gzipin {
		zin, err := gzip.NewReader(f)
		if err != nil {
			return nil
		}
		in = bufio.NewReader(zin)
	}

	err = ParseTriples(c, in)
	if err != nil {
		log.Printf("ReadTriplesFile error %v", err)
		return err
	}

	return nil
}

func Now() int64 {
	return time.Now().UTC().UnixNano()
}

func NowStringMillis() string {
	return time.Now().Format(time.RFC3339Nano)[0:23] + "Z"
}

func (g *Graph) LoadTriplesFile(filename string, opts *Options, wait *sync.WaitGroup) {
	defer func() {
		if wait != nil {
			wait.Done()
		}
	}()

	c := make(chan *Triple)
	go func() {
		ReadTriplesFile(c, filename)
	}()

	batchSize := 1000
	if n, ok := opts.IntKey("batch_size"); ok {
		batchSize = n
	}

	interval := 100000
	if n, ok := opts.IntKey("interval"); ok {
		interval = n
	}

	stats := false
	if b, ok := opts.BoolKey("stats"); ok {
		stats = b
	}

	i := 0
	then := Now()
	batch := make([]*Triple, 0, batchSize)
	wrote := 0
	problems := 0

	report := func() {
		time := Now()
		elapsed := time - then
		then = time
		rate := float64(interval) / float64(elapsed) * 1000000000.0
		fmt.Printf("load %s %s %012d %f %012d %06d %012d\n", filename, NowStringMillis(), i, rate, wrote, problems, g.GetWrites())
		if stats {
			fmt.Printf("%s\n", g.GetStats())
		}
	}
	for t := range c {
		if t == nil {
			log.Printf("nil triple")
			continue
		}

		if t == nil {
			break
		}
		i++
		if batchSize == len(batch) {
			err := g.WriteIndexedTriples(batch, nil)
			if err != nil {
				problems++
				fmt.Printf("ERROR: %v %d at %d\n", err, problems, i)
				for j, bad := range batch {
					ss := bad.Strings()
					s := ss[0]
					p := ss[1]
					o := ss[2]
					fmt.Printf("PROBLEM %d '%s' '%s' '%s'\n", j, s, p, o)
				}
			} else {
				wrote += len(batch)
			}
			batch = batch[0:0]
		}
		batch = append(batch, t)

		if i%interval == 0 {
			report()
		}

	}
	g.WriteIndexedTriples(batch, nil)
	report()

}
