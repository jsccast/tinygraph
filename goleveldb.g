package main

import (
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
	"github.com/syndtr/goleveldb/leveldb/iterator"
)

type Graph struct {
	db *leveldb.DB
	opts *opt.Options
	wopts *opt.WriteOptions
	ropts *opt.ReadOptions
}

func (g *Graph) Open(path string, opts *opt.Options) error {
	db, err := leveldb.OpenFile(path, opts)
	if err != nil {
		return err
	}
	g.db = db
	return nil
}

func (g *Graph) Close() error {
	if g.db == nil {
		return fmt.Errorf("Graph isn't open")
	}
	err := g.db.Close()
	if err != nil {
		return err
	}
	g.db = nil
	return nil
}

func (g *Graph) WriteBatch(triples []Triple, opts *opt.WriteOptions) error {
	batch := new(leveldb.Batch)
	for _,triple := range triples {
		batch.Put(triple.Key(), triple.Val())
	}
	return g.db.Write(batch, opts)
}

func IndexedTripleFromBytes(index Index, bs []byte, v []byte) *Triple {
	triple := TripleFromBytes(bs[1:], v)
	return triple
}

func withIndex(index Index, k []byte) []byte {
	bs := make([]byte, 0, len(k) + 1)
	bs = append(bs, byte(index))
	bs = append(bs, k...)
	return bs
}

func (g *Graph) IndexTriple(index Index, triple *Triple, opts *opt.WriteOptions) error {
	return g.db.Put(withIndex(index, triple.Key()), triple.Val(), opts)
}


func (g *Graph) NewIterator(on *Triple, opts *opt.ReadOptions) iterator.Iterator {
	span := util.Range{on.Prefix(), on.EndPrefix()}
	fmt.Printf("span %v\n", span)
	i := g.db.NewIterator(&span, opts)
	return i
}

func (g *Graph) NewIndexIterator(index Index, on *Triple, opts *opt.ReadOptions) iterator.Iterator {
	span := util.Range{withIndex(index, on.Prefix()), withIndex(index, on.EndPrefix())}
	// fmt.Printf("span %v %v\n", span, *on)
	i := g.db.NewIterator(&span, opts)
	return i
}

func (g *Graph) WriteTriple(triple *Triple, opts *opt.WriteOptions) error {
	return g.db.Put(triple.Key(), triple.Val(), opts)
}

func (g *Graph) WriteIndexedTriple(triple *Triple, opts *opt.WriteOptions) error {
	batch := new(leveldb.Batch)
	v := triple.Val()
	// ToDo: Optimize
	batch.Put(withIndex(SPO, triple.Copy().Permute(SPO).Key()), v)
	batch.Put(withIndex(OPS, triple.Copy().Permute(OPS).Key()), v)
	batch.Put(withIndex(PSO, triple.Copy().Permute(PSO).Key()), v)
	return g.db.Write(batch, opts)
}

func (g *Graph) WriteIndexedTriples(triples []*Triple, opts *opt.WriteOptions) error {
	batch := new(leveldb.Batch)
	for _, triple := range triples {
		v := triple.Val()
		// ToDo: Optimize
		batch.Put(withIndex(SPO, triple.Copy().Permute(SPO).Key()), v)
		batch.Put(withIndex(OPS, triple.Copy().Permute(OPS).Key()), v)
		batch.Put(withIndex(PSO, triple.Copy().Permute(PSO).Key()), v)
	}
	return g.db.Write(batch, opts)
}

func (g *Graph) Scan(index Index, on *Triple, opts *opt.ReadOptions) []Triple {
	acc := make([]Triple, 0, 64)
	i := g.NewIndexIterator(index, on, nil)
	for i.Next() {
		triple := IndexedTripleFromBytes(index, i.Key(), i.Value()).Permute(index)
		acc = append(acc, *triple)
	}
	i.Release()
	return acc
}

type TripleFun func(*Triple) bool

func (g *Graph) Do(index Index, on *Triple, opts *opt.ReadOptions, f TripleFun) error {
	i := g.NewIndexIterator(index, on, nil)
	for i.Next() {
		if !f(IndexedTripleFromBytes(index, i.Key(), i.Value()).Permute(index)) {
			break
		}
	}
	i.Release()
	return nil
}

func main() {
	g := Graph{}
	err := g.Open("g", &opt.Options{})
	if err != nil { panic(err) }

	err = g.WriteIndexedTriple(TripleFromStrings("I","liked","salad","today"), nil)
	err = g.WriteIndexedTriple(TripleFromStrings("I","ate","chips","today"), nil)
	err = g.WriteIndexedTriple(TripleFromStrings("I","sold","fruit","today"), nil)
	err = g.WriteIndexedTriple(TripleFromStrings("I","love","beer3","today"), nil)
	
	for n := 0; n < 5; n++ {
		beer := fmt.Sprintf("beer%d", n)
		err = g.WriteIndexedTriple(TripleFromStrings("I","like",beer,"today"), nil)
		if err != nil { panic(err) }
	}

	for n := 0; n < 5; n++ {
		tacos := fmt.Sprintf("tacos%d", n)
		err = g.WriteIndexedTriple(TripleFromStrings("I","love",tacos,"today"), nil)
		if err != nil { panic(err) }
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

	if false {
		g.LoadTriplesFromFile("some.nt")
	}

	fmt.Println("wordnet")
	g.Do(SPO, TripleFromStrings("100002452-n", "ontology#hyponym"), nil, PrintTriple)

	// rdf-schema#label

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

	err = g.Close()
	if err != nil { panic(err) }
}

