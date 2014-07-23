# A tiny graph database

Goal: A very fast and efficient graph database that handle billions of
vertexes and billions of edges on single machine.

Motivation: Semantic-social movie recommendations, but that's another
story.

The code is tiny, but the data can be single-machine big.

As a proof-of-concept, we load
[all of Freebase](https://developers.google.com/freebase/data), which
is currently 2.6B triples and 350GB uncompressed.  See below for
details.


For storage, the code currently requires
[`rocksdb`](http://rocksdb.org/), but many other back ends (e.g.,
[`levigo`](https://github.com/jmhodges/levigo) or
[`goleveldb`](https://github.com/syndtr/goleveldb)) would be easy to
support.

Also see [Cayley](https://github.com/google/cayley).  I have a
[fork](http://github.csv.comcast.com/jsteph206/cayley) that supports
[`rocksdb`](http://rocksdb.org/).  I wrote Tinygraph primarily because
I wanted to avoid Cayley's string interning (and reference counting)
and I didn't want to rewrite by rocksdb support for Cayley to store
strings directly.  (The latter probably would be easy.)

## Status

Highly experimental.

ToDo:

1. Logging.
2. Test cases.
3. Docs (especially configuration).
4. Document Wordnet load/use.


## Query language

Inspired in part by
[Cayley's Grelim-like language](https://github.com/google/cayley/blob/master/docs/GremlinAPI.md),
we have a simple, Javascript-based query language.  (Cayley uses
[`github.com/robertkrimen/otto`](github.com/robertkrimen/otto) like we
do
[elsewhere](https://github.com/google/cayley/blob/master/docs/GremlinAPI.md).)
However, unlike Cayley, we base this language on Go functions that can
be used from Go applications.

```Javascript
g = G.Open('config.json')
G.Out("p1").Walk(g, G.Vertex("a")).Collect()[0][0].ToStrings()
```

ToDo: Document.

## WordNet

It's easy to load [WordNet RDF](http://wordnet-rdf.princeton.edu/).

ToDo: Say how.  (Basically just set `triples_file` in `config.json`.)

It takes 4.5 seconds to load about 1M WordNet triples.

Example query:

```Javascript
paths = G.In("rdf-schema#label").Out(G.Bs("ontology#hyponym")).Out(G.Bs("rdf-schema#label")).Walk(g, G.Vertex("virus")).Collect();
for (var i=0; i<paths.length; i++) { console.log(paths[i][2].ToStrings()[2]); }
```

gives

```
arbovirus
arbovirus
phage
phage
plant virus
animal virus
slow virus
tumor virus
vector
```


## Freebase

Currently I'm using this code to load all of
[Freebase](https://developers.google.com/freebase/data).

### Machine
```
RAM        64GB
cores      24 (with HT)
model name Six-Core AMD Opteron(tm) Processor 8435
cpu MHz    2593.770
cache size 512 KB
disks      7xHDD (?)
```

### Tinygraph/rocksdb config

```Javascript
{
  "allow_mmap_reads": false,
  "allow_mmap_writes": false,
  "allow_os_buffer": true,
  "background_threads": 12,
  "batch_size": 12000,
  "block_size": 65536,
  "bytes_per_sync_power": 25,
  "cache_size": 3.3554432e+07,
  "compression": "snappy",
  "disable_data_sync": true,
  "disable_wal": true,
  "high_priority_background_threads": 12,
  "increase_parallelism": 12,
  "initial_compaction": false,
  "level0_num_file_compaction_trigger": 24,
  "log_level": 10,
  "max_background_compactions": 8,
  "max_background_flushes": 8,
  "max_open_files": 512,
  "max_write_buffer_number": 4,
  "min_write_buffer_number_to_merge": 4,
  "num_levels": 6,
  "paranoid_checks": false,
  "stats_dump_period": 300,
  "stats_loop": true,
  "sync": false,
  "target_file_size_base_power": 29,
  "target_file_size_multiplier": 4,
  "triples_file": "/dev/shm/rocks/in/1,/dev/shm/rocks/in/2,/dev/shm/rocks/in/3,/dev/shm/rocks/in/4,/dev/shm/rocks/in/5",
  "wal_dir": "/dev/shm/rocks",
  "write_buffer_size_power": 28
 }
```

### Processing

```
lines     2 638 544 493
bytes   356 018 834 809
start   2014-07-20T22:55:13.279Z
done    2014-07-20T06:57:23.543Z
elapsed 15:58:50
keys    Should be 8B, but still verifying (at 2 386 769 886, which is strange)
disk    89 943 764 K
```

### Rocksdb levels

```
Level Files Size(MB)
--------------------
  L0     6      478 
  L1     0        0 
  L2     2     1026 
  L3     7     1524 
  L4     4     7399 
  L5     3    77283 
 Sum    22    87710 
```

### Graphs


![fb-keys-over-time.png](images/fb-keys-over-time.png)

![fb-mean-triple-rate-over-time.png](images/fb-mean-triple-rate-over-time.png)

![fb-mean-triple-rate-over-time-zoom.png](images/fb-mean-triple-rate-over-time-zoom.png)

![fb-gc-time-over-time.png](images/fb-gc-time-over-time.png)

![fb-gc-time-zoom.png](images/fb-gc-time-zoom.png)


### Example query

```Go
g.Do(SPO, &Triple{[]byte("http://rdf.freebase.com/ns/m.0h55n27"), nil, nil, nil}, nil, ... )
```

```
next [http://rdf.freebase.com/ns/m.0h55n27 http://rdf.freebase.com/key/wikipedia.en America_ebolavirus ]
next [http://rdf.freebase.com/ns/m.0h55n27 http://rdf.freebase.com/key/wikipedia.en EBOV-R ]
next [http://rdf.freebase.com/ns/m.0h55n27 http://rdf.freebase.com/key/wikipedia.en Ebola_Reston ]
next [http://rdf.freebase.com/ns/m.0h55n27 http://rdf.freebase.com/key/wikipedia.en REBOV ]
next [http://rdf.freebase.com/ns/m.0h55n27 http://rdf.freebase.com/key/wikipedia.en RESTV ]
next [http://rdf.freebase.com/ns/m.0h55n27 http://rdf.freebase.com/key/wikipedia.en Reston_Ebola_virus ]
next [http://rdf.freebase.com/ns/m.0h55n27 http://rdf.freebase.com/key/wikipedia.en Reston_ebolavirus ]
next [http://rdf.freebase.com/ns/m.0h55n27 http://rdf.freebase.com/key/wikipedia.en Reston_virus ]
next [http://rdf.freebase.com/ns/m.0h55n27 http://rdf.freebase.com/key/wikipedia.en Virginia_ebolavirus ]
next [http://rdf.freebase.com/ns/m.0h55n27 http://rdf.freebase.com/key/wikipedia.en_id 33041857 ]
next [http://rdf.freebase.com/ns/m.0h55n27 http://rdf.freebase.com/key/wikipedia.en_title Reston_virus ]
next [http://rdf.freebase.com/ns/m.0h55n27 http://rdf.freebase.com/ns/common.topic.article http://rdf.freebase.com/ns/m.0h55n2c ]
next [http://rdf.freebase.com/ns/m.0h55n27 http://rdf.freebase.com/ns/common.topic.description Reston virus was first described in 1990 as a new "strain" of Ebola virus, a result of mutation from Ebola virus. It is the single member o\
f the species Reston ebolavirus, which is included into the genus Ebolavirus, family Filoviridae, order Mononegavirales. Reston virus is named after Reston, Virginia, US, where the virus was first discovered.
RESTV was discovered in crab-eating macaques from Hazleton Laboratories in 1989. This attracted significant media attention due to the proximity of Reston to the Washington, DC metro area, and the lethality of a closely related Ebola \
virus. Despite its status as a level-4 organism, Reston virus is non-pathogenic to humans, though hazardous to monkeys; the perception of its lethality was confounded due to the monkey's coinfection with Simian hemorrhagic fever virus\
. ]
next [http://rdf.freebase.com/ns/m.0h55n27 http://rdf.freebase.com/ns/common.topic.image http://rdf.freebase.com/ns/m.059jkjn ]
next [http://rdf.freebase.com/ns/m.0h55n27 http://rdf.freebase.com/ns/common.topic.notable_for http://rdf.freebase.com/ns/g.1256ncwfc ]
next [http://rdf.freebase.com/ns/m.0h55n27 http://rdf.freebase.com/ns/common.topic.notable_types http://rdf.freebase.com/ns/m.03sp3gw ]
next [http://rdf.freebase.com/ns/m.0h55n27 http://rdf.freebase.com/ns/common.topic.topic_equivalent_webpage http://en.wikipedia.org/wiki/Reston_virus ]
next [http://rdf.freebase.com/ns/m.0h55n27 http://rdf.freebase.com/ns/common.topic.topic_equivalent_webpage http://en.wikipedia.org/wiki/index.html?curid=33041857 ]
next [http://rdf.freebase.com/ns/m.0h55n27 http://rdf.freebase.com/ns/type.object.key /wikipedia/en/America_ebolavirus ]
next [http://rdf.freebase.com/ns/m.0h55n27 http://rdf.freebase.com/ns/type.object.key /wikipedia/en/EBOV-R ]
next [http://rdf.freebase.com/ns/m.0h55n27 http://rdf.freebase.com/ns/type.object.key /wikipedia/en/Ebola_Reston ]
next [http://rdf.freebase.com/ns/m.0h55n27 http://rdf.freebase.com/ns/type.object.key /wikipedia/en/REBOV ]
next [http://rdf.freebase.com/ns/m.0h55n27 http://rdf.freebase.com/ns/type.object.key /wikipedia/en/RESTV ]
next [http://rdf.freebase.com/ns/m.0h55n27 http://rdf.freebase.com/ns/type.object.key /wikipedia/en/Reston_Ebola_virus ]
next [http://rdf.freebase.com/ns/m.0h55n27 http://rdf.freebase.com/ns/type.object.key /wikipedia/en/Reston_ebolavirus ]
next [http://rdf.freebase.com/ns/m.0h55n27 http://rdf.freebase.com/ns/type.object.key /wikipedia/en/Reston_virus ]
next [http://rdf.freebase.com/ns/m.0h55n27 http://rdf.freebase.com/ns/type.object.key /wikipedia/en/Virginia_ebolavirus ]
next [http://rdf.freebase.com/ns/m.0h55n27 http://rdf.freebase.com/ns/type.object.key /wikipedia/en_id/33041857 ]
next [http://rdf.freebase.com/ns/m.0h55n27 http://rdf.freebase.com/ns/type.object.key /wikipedia/en_title/Reston_virus ]
next [http://rdf.freebase.com/ns/m.0h55n27 http://rdf.freebase.com/ns/type.object.name Reston virus ]
next [http://rdf.freebase.com/ns/m.0h55n27 http://rdf.freebase.com/ns/type.object.type http://rdf.freebase.com/ns/common.topic ]
next [http://rdf.freebase.com/ns/m.0h55n27 http://rdf.freebase.com/ns/type.object.type http://rdf.freebase.com/ns/medicine.disease_cause ]
next [http://rdf.freebase.com/ns/m.0h55n27 http://www.w3.org/1999/02/22-rdf-syntax-ns#type http://rdf.freebase.com/ns/common.topic ]
next [http://rdf.freebase.com/ns/m.0h55n27 http://www.w3.org/1999/02/22-rdf-syntax-ns#type http://rdf.freebase.com/ns/medicine.disease_cause ]
next [http://rdf.freebase.com/ns/m.0h55n27 http://www.w3.org/2000/01/rdf-schema#label Reston virus ]
```


### Notes

I sometimes [listened](images/load.mp3) to spot rates while doing other things.  Funny.
