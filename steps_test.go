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
		t.Fatalf("2 Expected %d paths but got %d", 1, len(paths))
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
		t.Fatalf("3 Expected %d paths but got %d", 1, len(paths))
	}
	if string(paths[0][3].O) != "i" {
		t.Error("3 Expected %s but got %s", "i", paths[0][3].O)
	}

	// 4
	paths = AllOut().
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

	AllOut().
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

	g.Close()

}

func TestDo(t *testing.T) {

	// For now, we just pile stuff into this one function.
	// ToDo: Don't do that.

	g, _ := GetGraph("config.test")

	g.WriteIndexedTriple(TripleFromStrings("aaa", "p1", "bbb", "today"), nil)
	g.WriteIndexedTriple(TripleFromStrings("aaa", "p1", "fff", "today"), nil)
	g.WriteIndexedTriple(TripleFromStrings("aaa", "p5", "jjj", "today"), nil)
	g.WriteIndexedTriple(TripleFromStrings("bbb", "p2", "ccc", "today"), nil)
	g.WriteIndexedTriple(TripleFromStrings("ccc", "p3", "ddd", "today"), nil)
	g.WriteIndexedTriple(TripleFromStrings("ccc", "p3", "eee", "today"), nil)
	g.WriteIndexedTriple(TripleFromStrings("ggg", "p4", "ccc", "today"), nil)
	g.WriteIndexedTriple(TripleFromStrings("ggg", "p1", "hhh", "today"), nil)
	g.WriteIndexedTriple(TripleFromStrings("ggg", "p1", "iii", "today"), nil)

	g.DoAll(nil, 100, func(t *Triple) bool {
		fmt.Printf("triple %v\n", t.String())
		return true
	})

	fmt.Println("DoS")

	g.DoVertexes(nil, 30, func(s []byte) bool {
		Out([]byte("p1")).Walk(g, s).Print()
		return true
	})

	i := g.NewVertexIterator()
	for {
		v, ok := i.Next()
		if !ok {
			break
		}
		fmt.Printf("v %s\n", v)
	}

	i = g.NewVertexIterator()
	for {
		v := i.NextVertex()
		if v == "" {
			break
		}
		fmt.Printf("nv %s\n", v)
	}

	g.Close()
}
