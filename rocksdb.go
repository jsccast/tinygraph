package main

import (
	"bytes"
	"fmt"
	rocks "github.csv.comcast.com/jsteph206/gorocksdb"
	"sync/atomic"
)

type Graph struct {
	db     *rocks.DB
	opts   *rocks.Options
	wopts  *rocks.WriteOptions
	ropts  *rocks.ReadOptions
	writes uint64
}

func NewGraph(path string, opts *rocks.Options) (*Graph, error) {
	db, err := rocks.Open(path, opts)
	if err != nil {
		return nil, err
	}
	return &Graph{db, opts, nil, nil, uint64(0)}, nil
}

func (g *Graph) IncWrites(n uint64) uint64 {
	return atomic.AddUint64(&g.writes, n)
}

func (g *Graph) GetWrites() uint64 {
	return atomic.LoadUint64(&g.writes)
}

func (g *Graph) GetStats() string {
	return g.db.PropertyValue("rocksdb.stats") + fmt.Sprintf("\nnewwrites %d\n", g.GetWrites())
}

func (g *Graph) Close() error {
	if g.db == nil {
		return fmt.Errorf("Graph isn't open")
	}
	g.db.Close()
	g.db = nil
	return nil
}

func (g *Graph) WriteBatch(triples []Triple, opts *rocks.WriteOptions) error {
	if opts == nil {
		opts = g.wopts
	}
	batch := rocks.NewWriteBatch()
	for _, triple := range triples {
		batch.Put(triple.Key(), triple.Val())
	}
	err := g.db.Write(opts, batch)
	if err == nil {
		g.IncWrites(uint64(len(triples)))
	}
	return err
}

func IndexedTripleFromBytes(index Index, bs []byte, v []byte) *Triple {
	triple := TripleFromBytes(bs[1:], v)
	return triple
}

func withIndex(index Index, k []byte) []byte {
	bs := make([]byte, 0, len(k)+1)
	bs = append(bs, byte(index))
	bs = append(bs, k...)
	return bs
}

func (g *Graph) IndexTriple(index Index, triple *Triple, opts *rocks.WriteOptions) error {
	if opts == nil {
		opts = g.wopts
	}
	err := g.db.Put(opts, withIndex(index, triple.Key()), triple.Val())
	if err == nil {
		g.IncWrites(1)
	}
	return err
}

type IteratorState byte

const (
	Init = iota
	Active
	Done
)

type Iterator struct {
	i     *rocks.Iterator
	from  []byte
	to    []byte
	state IteratorState
}

func (i *Iterator) Key() []byte {
	return i.i.Key()
}

func (i *Iterator) Value() []byte {
	return i.i.Value()
}

func (i *Iterator) Release() {
	i.i.Close()
	i.state = Done
}

func (g *Graph) NewIndexIterator(index Index, on *Triple, opts *rocks.ReadOptions) *Iterator {
	if opts == nil {
		opts = g.ropts
	}
	from := withIndex(index, on.StartKey())
	to := withIndex(index, on.KeyPrefix())
	i := &Iterator{g.db.NewIterator(opts), from, to, Init}
	return i
}

func (i *Iterator) Next() bool {
	switch i.state {
	case Init:
		i.i.Seek(i.from)
		i.state = Active
	case Active:
		i.i.Next()
	case Done:
		return false
	}

	if !i.i.Valid() {
		i.state = Done
		return false
	}

	bs := i.i.Key()

	if !bytes.HasPrefix(bs, i.to) {
		i.state = Done
		return false
	}

	return true
}

func (g *Graph) WriteTriple(triple *Triple, opts *rocks.WriteOptions) error {
	if opts == nil {
		opts = g.wopts
	}
	err := g.db.Put(opts, triple.Key(), triple.Val())
	if err == nil {
		g.IncWrites(1)
	}
	return err
}

func (g *Graph) WriteIndexedTriple(triple *Triple, opts *rocks.WriteOptions) error {
	if opts == nil {
		opts = g.wopts
	}
	if opts == nil {
		panic(fmt.Errorf("nil wopts"))
	}
	batch := rocks.NewWriteBatch()
	v := triple.Val()
	// ToDo: Optimize
	batch.Put(withIndex(SPO, triple.Copy().Permute(SPO).Key()), v)
	batch.Put(withIndex(OPS, triple.Copy().Permute(OPS).Key()), v)
	batch.Put(withIndex(PSO, triple.Copy().Permute(PSO).Key()), v)
	err := g.db.Write(opts, batch)
	if err == nil {
		g.IncWrites(uint64(3))
	}
	return err
}

func (g *Graph) WriteIndexedTriples(triples []*Triple, opts *rocks.WriteOptions) error {
	if opts == nil {
		opts = g.wopts
	}
	batch := rocks.NewWriteBatch()
	for _, triple := range triples {
		// fmt.Printf("triple %v\n", triple.ToString())
		v := triple.Val()
		// ToDo: Optimize
		batch.Put(withIndex(SPO, triple.Copy().Permute(SPO).Key()), v)
		batch.Put(withIndex(OPS, triple.Copy().Permute(OPS).Key()), v)
		batch.Put(withIndex(PSO, triple.Copy().Permute(PSO).Key()), v)
	}
	err := g.db.Write(opts, batch)
	if err == nil {
		g.IncWrites(uint64(3 * len(triples)))
	}
	return err
}

func (g *Graph) Scan(index Index, on *Triple, opts *rocks.ReadOptions) []Triple {
	acc := make([]Triple, 0, 64)
	i := g.NewIndexIterator(index, on, opts)
	for i.Next() {
		triple := IndexedTripleFromBytes(index, i.Key(), i.Value()).Permute(index)
		acc = append(acc, *triple)
	}
	i.Release()
	return acc
}

type TripleFun func(*Triple) bool

func (g *Graph) Do(index Index, on *Triple, opts *rocks.ReadOptions, f TripleFun) error {
	i := g.NewIndexIterator(index, on, opts)
	for i.Next() {
		if !f(IndexedTripleFromBytes(index, i.Key(), i.Value()).Permute(index)) {
			break
		}
	}
	i.Release()
	return nil
}
