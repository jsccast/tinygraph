package main

import (
	"fmt"
	"testing"
)

func TestSteps(t *testing.T) {

	// For now, we just pile stuff into this one function.
	// ToDo: Don't do that.

	g, _ := GetGraph("config.test")

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
		return string(t.O) == "i"
	}

	// 1
	v := Vertex([]byte("a"))
	expect := []string{"b", "f", "j"}
	for i, path := range g.Walk(v, []*Stepper{AllOut()}).Collect() {
		o := path[0]
		if string(o.O) != expect[i] {
			t.Errorf("1 Expected %s at %d but got %s", expect[i], i, o.O)
		}
	}

	// 2
	v = Vertex([]byte("a"))
	paths := g.Walk(v, []*Stepper{
		Out([]byte("p1")),
		Out([]byte("p2")),
		In([]byte("p4")),
		Out([]byte("p1")),
		Has(has)}).Collect()
	if len(paths) != 1 {
		t.Fatal("2 Expected %d paths but got %d", 1, len(paths))
	}
	if string(paths[0][3].O) != "i" {
		t.Error("2 Expected %s but got %s", "i", paths[0][3].O)
	}

	// 3
	paths = Out([]byte("p1")).
		Out([]byte("p2")).
		In([]byte("p4")).
		Out([]byte("p1")).
		Has(has).
		Walk(g, v).
		Collect()
	if len(paths) != 1 {
		t.Fatal("3 Expected %d paths but got %d", 1, len(paths))
	}
	if string(paths[0][3].O) != "i" {
		t.Error("3 Expected %s but got %s", "i", paths[0][3].O)
	}

	// 4
	paths = Out(nil).
		Out([]byte("p2")).
		In([]byte("p4")).
		Out([]byte("p1")).
		Has(has).
		Walk(g, v).
		Collect()
	if len(paths) != 1 {
		t.Fatal("4 Expected %d paths but got %d", 1, len(paths))
	}
	if string(paths[0][3].O) != "i" {
		t.Error("4 Expected %s but got %s", "i", paths[0][3].O)
	}

	Out(nil).
		Do(func(path Path) { fmt.Printf("doing 0 %v\n", path.String()) }).
		Out([]byte("p2")).
		Do(func(path Path) { fmt.Printf("doing 1 %v\n", path.String()) }).
		In([]byte("p4")).
		Do(func(path Path) { fmt.Printf("doing 2 %v\n", path.String()) }).
		Out([]byte("p1")).
		Has(func(t Triple) bool { return string(t.O) == "i" }).
		Walk(g, v).
		Do(func(path Path) { fmt.Printf("got %v\n", path.String()) })

	Out([]byte("p1")).
		Walk(g, v).
		Print()

	fmt.Printf("collected %v\n", Out([]byte("p1")).
		Walk(g, v).
		Collect())
}