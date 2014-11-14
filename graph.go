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

// How to read and write triples.

import (
	"bytes"
	"fmt"
	"log"
	"sync/atomic"

	rocks "github.com/jsccast/rocksdb"
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

func (g *Graph) Compact() {
	log.Printf("starting initial compaction %s\n", NowStringMillis())
	ff := byte(0xff)
	r := rocks.Range{[]byte{}, []byte{ff, ff, ff, ff, ff, ff, ff, ff, ff}}
	g.db.CompactRange(r)
	log.Printf("completed initial compaction %s\n", NowStringMillis())
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
	var from []byte
	var to []byte
	if on == nil {
		zero := []byte{}
		from = withIndex(index, zero)
		to = withIndex(index, zero)
	} else {
		from = withIndex(index, on.StartKey())
		to = withIndex(index, on.KeyPrefix())
	}
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

func (g *Graph) DoAll(opts *rocks.ReadOptions, limit int, f TripleFun) error {
	i := g.NewIndexIterator(SPO, nil, opts)
	for i.Next() && 0 < limit {
		limit--
		if !f(IndexedTripleFromBytes(SPO, i.Key(), i.Value()).Permute(SPO)) {
			break
		}
	}
	i.Release()
	return nil
}

func inc(bs []byte) []byte { // Copies
	acc := make([]byte, len(bs))
	copy(acc, bs)

	i := len(bs) - 1
	end := len(bs)
	for 0 <= i {
		b := acc[i]
		if b < b+1 {
			acc[i]++
			break
		}
		// Overflow
		end--
		i--
	}
	if i < 0 {
		return []byte{}
	}

	return acc[0:end]
}

func (g *Graph) DoVertexes(opts *rocks.ReadOptions, limit int, f func([]byte) bool) error {
	// ToDo: Reimplement with iterators?
	if opts == nil {
		opts = g.ropts
	}

	i := g.db.NewIterator(opts)
	index := byte(SPO)
	at := []byte{index, 0}

	for 0 <= limit {
		limit--
		i.Seek(at)
		if !i.Valid() {
			break
		}
		k := i.Key()
		if k[0] != index {
			break
		}
		v := i.Value()
		t := TripleFromBytes(k[1:], v)
		s := t.S
		if !f(s) {
			break
		}
		s = inc(s)
		at = make([]byte, len(s)+1)
		copy(at[1:], s)
		at[0] = index
	}

	i.Close()
	return nil
}

type VertexIterator struct {
	i    *rocks.Iterator
	at   []byte
	done bool
}

func (g *Graph) NewVertexIterator() *VertexIterator {
	opts := g.ropts
	return &VertexIterator{g.db.NewIterator(opts), []byte{}, false}
}

func (i *VertexIterator) Next() ([]byte, bool) {
	index := byte(SPO)

	at := i.at
	if i.done {
		return nil, false
	}

	if len(at) == 0 {
		at = []byte{index, 0}
	}

	i.i.Seek(at)
	if !i.i.Valid() {
		i.Release()
		return nil, false
	}

	k := i.i.Key()
	if k[0] != index {
		i.Release()
		return nil, false
	}

	v := i.i.Value()
	t := TripleFromBytes(k[1:], v)
	s := t.S

	seek := inc(s)
	to := make([]byte, len(seek)+1)
	copy(to[1:], seek)
	to[0] = index
	i.at = to

	return s, true
}

// Really for Javscript.
func (i *VertexIterator) NextVertex() string {
	bs, ok := i.Next()
	if !ok {
		return ""
	}
	return string(bs)
}

func (i *VertexIterator) Release() {
	if !i.done {
		i.i.Close()
		i.done = true
	}
}
