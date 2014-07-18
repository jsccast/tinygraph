package main

import (
	"fmt"
	"os"
	"sync"
	"time"
)

func ReadTriplesFromFile(c chan *Triple, tripleFile string) error {
	f, err := os.Open(tripleFile)
	if err != nil {
		fmt.Printf("Couldn't open file %s: %v\n", tripleFile, err)
		return fmt.Errorf("Couldn't open file %s: %v", tripleFile, err)
	}

	ReadNQuadsFromReader(c, f)
	if err := f.Close(); err != nil {
		fmt.Printf("%v\n", err)
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

func (g *Graph) LoadTriplesFromFile(filename string, opts *Options, wait *sync.WaitGroup) {
	c := make(chan *Triple)
	go ReadTriplesFromFile(c, filename)
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
	for t := range c {
		i++
		if batchSize == len(batch) {
			g.WriteIndexedTriples(batch, nil)
			batch = batch[0:0]
		}
		batch = append(batch, t)

		if i%interval == 0 {
			time := Now()
			elapsed := time - then
			then = time
			rate := float64(interval) / float64(elapsed) * 1000000000.0
			fmt.Printf("load %s %s %012d %f\n", filename, NowStringMillis(), i, rate)
			if stats {
				fmt.Printf("%s\n", g.GetStats())
			}
		}

	}
	g.WriteIndexedTriples(batch, nil)

	if wait != nil {
		wait.Done()
	}
}
