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
var serve = flag.Bool("serve", false, "Start HTTPD server")
var onlyLang = flag.String("lang", "eng", "Only get these strings ('en' for Freebase; 'eng' for WordNet)")
var configFile = flag.String("config", "config.js", "Configuration file")
var sharedHttpVM = flag.Bool("sharevm", true, "Use a shared Javascript VM for the HTTP service")
var chanBufferSize = flag.Int("chanbuf", 16, "Traversal emission buffer")
var httpPort = flag.String("port", ":8080", "HTTP server port")
var gzipin = flag.Bool("gzip", false, "Input triple files are gzipped")
var ignoreSilently = flag.Bool("silent-ignore", true, "Don't report when ingoring a triple")

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
			ioutil.WriteFile("stats.log", bs, 0644)
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

	dirname := "tmp.db"
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
	g, config := GetGraph(*configFile)
	fmt.Println(g.GetStats())

	if b, ok := config.BoolKey("initial_compaction"); ok && b {
		CompactEverything(g)
		fmt.Println(g.GetStats())
	}

	if b, ok := config.BoolKey("stats_loop"); ok && b {
		WriteStatsLoop(g)
	}

	wait := sync.WaitGroup{}
	for _, filename := range strings.Split(*filesToLoad, ",") {
		filename = strings.TrimSpace(filename)
		fmt.Printf("loading triples: %s\n", filename)
		wait.Add(1)
		go g.LoadTriplesFile(filename, config, &wait)
		// Stagger the threads a little.
		time.Sleep(1 * time.Second)
	}
	wait.Wait()

	fmt.Println(g.GetStats())

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
