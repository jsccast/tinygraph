package main

import (
	"fmt"
	"testing"
	. "github.csv.comcast.com/jsteph206/tinygraph"
)

func TinyTest(g *Graph) {

	// For now, pile a bunch of tests in here.
	// ToDo: Not that.

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

func TestTinygraph(t *testing.T) {
	g, _ := GetGraph(*configFile)
	TinyTest(g)
	g.Close()
}
