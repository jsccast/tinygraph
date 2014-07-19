package main

import (
	"fmt"
	rocks "github.com/DanielMorsing/rocksdb"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

func main() {
	if os.Getenv("GOMAXPROCS") == "" {
		n := runtime.NumCPU()
		fmt.Printf("Setting GOMAXPROCS to %d\n", n)
		runtime.GOMAXPROCS(n)
	} else {
		fmt.Printf("GOMAXPROCS is %v\n", os.Getenv("GOMAXPROCS"))
	}

	config, err := LoadOptions("config.json")
	if err != nil {
		panic(err)
	}

	opts := RocksOpts(config)
	opts.SetCreateIfMissing(true)
	opts.SetErrorIfExists(false)
	g, err := NewGraph("g", opts)

	if err != nil {
		panic(err)
	}

	g.wopts = RocksWriteOpts(config)
	g.ropts = RocksReadOpts(config)

	fmt.Println(g.GetStats())
	if b, ok := config.BoolKey("initial_compaction"); ok {
		if b {
			fmt.Printf("starting initial compaction %s\n", NowStringMillis())
			ff := byte(0xff)
			r := rocks.Range{[]byte{}, []byte{ff, ff, ff, ff, ff, ff, ff, ff, ff}}
			g.db.CompactRange(r)
			fmt.Printf("completed initial compaction %s\n", NowStringMillis())
			fmt.Println(g.GetStats())
		}
	}

	go func() {
		for {
			time.Sleep(5 * time.Second)
			bs := []byte(g.GetStats())
			ioutil.WriteFile("stats.txt", bs, 0644)
		}
	}()

	if filenames, ok := config.StringKey("triples_file"); ok {
		wait := sync.WaitGroup{}
		for _, filename := range strings.Split(filenames, ",") {
			filename = strings.TrimSpace(filename)
			fmt.Printf("loading triples: %s\n", filename)
			wait.Add(1)
			go g.LoadTriplesFromFile(filename, config, &wait)
			// Stagger the threads a little.
			time.Sleep(1 * time.Second)
		}
		wait.Wait()
	}

	fmt.Println(g.GetStats())

	Bar(g)

	err = g.Close()
	if err != nil {
		panic(err)
	}
}

func Foo(g *Graph) {
	t := Triple{nil, nil, []byte("http://rdf.freebase.com/ns/m.0dw28j5"), nil}
	c := g.Walk(t, []Stepper{
		In([]byte("http://rdf.freebase.com/ns/type.type.instance")),
	})

	for {
		x := <-*c
		if x == nil {
			break
		}
		fmt.Printf("got %v\n", x)
	}
}

func Bar(g *Graph) {
	on := Triple{[]byte("http://rdf.freebase.com/ns/m.0h55n27"), nil, nil, nil}
	i := g.NewIndexIterator(SPO, &on, nil)
	limit := 100
	for i.Next() {
		if limit == 0 {
			break
		}
		limit--
		fmt.Printf("next %v\n", IndexedTripleFromBytes(SPO, i.Key(), i.Value()).ToStrings())
	}
}

func TinyTest(g *Graph) {
	err := g.WriteIndexedTriple(TripleFromStrings("I", "liked", "salad", "today"), nil)
	err = g.WriteIndexedTriple(TripleFromStrings("I", "ate", "chips", "today"), nil)
	err = g.WriteIndexedTriple(TripleFromStrings("I", "sold", "fruit", "today"), nil)
	err = g.WriteIndexedTriple(TripleFromStrings("I", "love", "beer3", "today"), nil)

	for n := 0; n < 5; n++ {
		beer := fmt.Sprintf("beer%d", n)
		err = g.WriteIndexedTriple(TripleFromStrings("I", "like", beer, "today"), nil)
		if err != nil {
			panic(err)
		}
	}

	for n := 0; n < 5; n++ {
		tacos := fmt.Sprintf("tacos%d", n)
		err = g.WriteIndexedTriple(TripleFromStrings("I", "love", tacos, "today"), nil)
		if err != nil {
			panic(err)
		}
	}

	fmt.Println("I like")
	g.Do(SPO, TripleFromStrings("I", "like"), nil, PrintTriple)

	fmt.Println("I love")
	g.Do(SPO, TripleFromStrings("I", "love"), nil, PrintTriple)

	fmt.Println("I")
	g.Do(SPO, TripleFromStrings("I"), nil, PrintTriple)

	fmt.Println("beer3")
	g.Do(OPS, TripleFromStrings("beer3"), nil, PrintTriple)

	fmt.Println("love")
	g.Do(PSO, TripleFromStrings("love"), nil, PrintTriple)

	fmt.Println("wordnet")
	g.Do(SPO, TripleFromStrings("100002452-n", "ontology#hyponym"), nil, PrintTriple)

	/*

		{
			c := make(chan *Triple)
			f := func(t *Triple) bool {
				c <- t
				g.Do(SPO, TripleFromStrings(string(t.O), "ontology#hyponym"), nil,
					func (u *Triple) bool {
						c <- u
						return true
					})
				return true
			}
			go func() {
				g.Do(SPO, TripleFromStrings("100002452-n", "ontology#hyponym"), nil, f)
				c <- nil
			}()
			for {
				t := <- c
				if t == nil {
					break
				}
				g.Do(SPO, TripleFromStrings(string(t.O), "rdf-schema#label"), nil, PrintTriple)
			}
		}

	*/

	g.WriteIndexedTriple(TripleFromStrings("a", "p1", "b", "today"), nil)
	g.WriteIndexedTriple(TripleFromStrings("a", "p1", "f", "today"), nil)
	g.WriteIndexedTriple(TripleFromStrings("a", "p5", "j", "today"), nil)
	g.WriteIndexedTriple(TripleFromStrings("b", "p2", "c", "today"), nil)
	g.WriteIndexedTriple(TripleFromStrings("c", "p3", "d", "today"), nil)
	g.WriteIndexedTriple(TripleFromStrings("c", "p3", "e", "today"), nil)
	g.WriteIndexedTriple(TripleFromStrings("g", "p4", "c", "today"), nil)
	g.WriteIndexedTriple(TripleFromStrings("g", "p1", "h", "today"), nil)
	g.WriteIndexedTriple(TripleFromStrings("g", "p1", "i", "today"), nil)

	has := func(t Triple) bool {
		return true
	}

	t := Triple{nil, nil, []byte("a"), nil}
	c := g.Walk(t, []Stepper{
		Out([]byte("p1")).Emitter(func(ts []Triple) interface{} { return "p1:" + string(last(ts).O) }),
		Out([]byte("p2")),
		In([]byte("p4")),
		Out([]byte("p1")),
		Has(has).Emitter(func(ts []Triple) interface{} { return PathToString(ts) + " last" })})
	for {
		x := <-*c
		if x == nil {
			break
		}
		fmt.Printf("got %v\n", x)
	}
}
