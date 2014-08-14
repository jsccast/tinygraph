package main

import (
	"flag"
	"fmt"
	rocks "github.csv.comcast.com/jsteph206/gorocksdb"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

var filesToLoad = flag.String("load", "", "Files to load")
var repl = flag.Bool("repl", false, "Run REPL")
var onlyLang = flag.String("lang", "en", "Only get these strings ('en' for Freebase; 'eng' for WordNet)")

func RationalizeMaxProcs() {
	if os.Getenv("GOMAXPROCS") == "" {
		n := runtime.NumCPU()
		fmt.Printf("Setting GOMAXPROCS to %d\n", n)
		runtime.GOMAXPROCS(n)
	} else {
		fmt.Printf("GOMAXPROCS is %v\n", os.Getenv("GOMAXPROCS"))
	}
}

func CompactEverything(g *Graph) {
	fmt.Printf("starting initial compaction %s\n", NowStringMillis())
	ff := byte(0xff)
	r := rocks.Range{[]byte{}, []byte{ff, ff, ff, ff, ff, ff, ff, ff, ff}}
	g.db.CompactRange(r)
	fmt.Printf("completed initial compaction %s\n", NowStringMillis())
}

func WriteStatsLoop(g *Graph) {
	go func() {
		for {
			time.Sleep(10 * time.Second)
			bs := []byte(g.GetStats())
			ioutil.WriteFile("stats.txt", bs, 0644)
		}
	}()
}

func GetGraph(configFilename string) (*Graph, *Options) {
	config, err := LoadOptions(configFilename)
	if err != nil {
		panic(err)
	}

	opts := RocksOpts(config)
	opts.SetCreateIfMissing(true)
	opts.SetErrorIfExists(false)

	dirname := "tmprocks"
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

func Load() {
	g, config := GetGraph("config.json")
	fmt.Println(g.GetStats())

	if b, ok := config.BoolKey("initial_compaction"); ok && b {
		CompactEverything(g)
		fmt.Println(g.GetStats())
	}

	if b, ok := config.BoolKey("stats_loop"); ok && b {
		WriteStatsLoop(g)
	}

	// filenames,given := config.StringKey("triples_file")

	wait := sync.WaitGroup{}
	for _, filename := range strings.Split(*filesToLoad, ",") {
		filename = strings.TrimSpace(filename)
		fmt.Printf("loading triples: %s\n", filename)
		wait.Add(1)
		go g.LoadTriplesFromFile(filename, config, &wait)
		// Stagger the threads a little.
		time.Sleep(1 * time.Second)
	}
	wait.Wait()

	fmt.Println(g.GetStats())
	// fmt.Printf("Freebase check: %v\n", FreebaseCheck(g))

	// TinyTest(g)
	// StepsTest(g)
	// FreebaseCheck(g)

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
	if *repl {
		REPL()
	}
}

func DoPrint(g *Graph, index Index, label string, s string) bool {
	limit := 100
	found := 0
	g.Do(index, &Triple{[]byte(s), nil, nil, nil}, nil,
		func(t *Triple) bool {
			fmt.Printf("%s %v\n", label, t.ToStrings())
			found++
			limit--
			if limit == 0 {
				return false
			}
			return true
		})
	return 0 < found
}

func FreebaseCheck(g *Graph) bool {
	DoPrint(g, SPO, "SPO m.0h55n27", "http://rdf.freebase.com/ns/m.0h55n27")
	DoPrint(g, SPO, "SPO m.03sp3gw", "http://rdf.freebase.com/ns/m.03sp3gw")
	DoPrint(g, OPS, "OPS m.03sp3gw", "http://rdf.freebase.com/ns/m.03sp3gw")
	DoPrint(g, SPO, "SPO g.1256ncwfc", "http://rdf.freebase.com/ns/g.1256ncwfc")
	return DoPrint(g, OPS, "OPS g.1256ncwfc", "http://rdf.freebase.com/ns/g.1256ncwfc")
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

	g.WriteIndexedTriple(TripleFromStrings("a", "p1", "b", "today"), nil)
	g.WriteIndexedTriple(TripleFromStrings("a", "p1", "f", "today"), nil)
	g.WriteIndexedTriple(TripleFromStrings("a", "p5", "j", "today"), nil)
	g.WriteIndexedTriple(TripleFromStrings("b", "p2", "c", "today"), nil)
	g.WriteIndexedTriple(TripleFromStrings("c", "p3", "d", "today"), nil)
	g.WriteIndexedTriple(TripleFromStrings("c", "p3", "e", "today"), nil)
	g.WriteIndexedTriple(TripleFromStrings("g", "p4", "c", "today"), nil)
	g.WriteIndexedTriple(TripleFromStrings("g", "p1", "h", "today"), nil)
	g.WriteIndexedTriple(TripleFromStrings("g", "p1", "i", "today"), nil)

}
