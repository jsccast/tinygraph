package main

import (
	"fmt"
	"sync"
)

func labels(g *Graph, id string) []string {
	rel := []byte("http://www.w3.org/2000/01/rdf-schema#label")
	paths := Out(rel).Walk(g, Vertex(id)).Collect()
	acc := make([]string, 0, len(paths)+1)
	acc = append(acc, id)
	for _, path := range paths {
		acc = append(acc, path[0].Strings()[2])
	}
	return acc
}

func follow(g *Graph, rel []byte, id string, recursive bool, reverse bool, depth int, level int, memo *Memo) string {

	if have, seen := memo.Get(id); seen {
		return have
	}

	var paths []Path
	if reverse {
		paths = In(rel).Walk(g, Vertex(id)).Collect()
	} else {
		paths = Out(rel).Walk(g, Vertex(id)).Collect()
	}
	acc := "[["
	for i, term := range labels(g, id) {
		if 0 < i {
			acc += ","
		}
		acc += fmt.Sprintf(`"%s"`, term)
	}
	acc += "]"

	for _, path := range paths {
		h := path[0].Strings()[2]
		if recursive {
			if level <= depth {
				acc += "," + follow(g, rel, h, recursive, reverse, depth, level+1, memo)
			} else {
				acc += fmt.Sprintf(`,"%s"`, h)
			}
		}
	}
	acc += "]"
	memo.Set(id, acc)
	return acc
}

func recurse(g *Graph, rel string, reverse bool, depth int, id string, memo *Memo) {
	line := follow(g, []byte(rel), id, true, reverse, depth, 0, memo)
	fmt.Println(line)
}

func hypernyms(g *Graph, id string, memo *Memo) {
	recurse(g, "http://wordnet-rdf.princeton.edu/ontology#hypernym", false, 100, id, memo)
}

type Memo struct {
	sync.RWMutex
	state map[string]string
	hits  int
}

func (m *Memo) Get(id string) (string, bool) {
	m.RLock()
	have, seen := m.state[id]
	if seen {
		m.hits++
	}
	m.RUnlock()
	return have, seen
}

func (m *Memo) Set(id string, val string) {
	m.Lock()
	m.state[id] = val
	m.Unlock()
}

func allHypernyms(g *Graph, limit int, concurrency int) {
	i := g.NewVertexIterator()
	memo := &Memo{}
	memo.state = make(map[string]string)
	var wait sync.WaitGroup
	available := concurrency
	for limit < 0 || 0 < limit {
		v := i.NextVertex()
		if len(v) == 0 {
			break
		}
		for {
			if 0 < available {
				wait.Add(1)
				go func(v string) {
					hypernyms(g, v, memo)
					wait.Done()
				}(v)
				available--
				limit--
				break
			}
			wait.Wait()
			available = concurrency
		}
	}
	wait.Wait()
	i.Release()
	fmt.Printf("memo hits: %d entries: %d\n", memo.hits, len(memo.state))
}

func (e *Env) Example(g *Graph, limit int64, concurrency int64) {
	allHypernyms(g, int(limit), int(concurrency))
}
