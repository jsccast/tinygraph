# A tiny graph database

Goal: A simple and relatively efficient graph data store that can
handle billions of vertexes on single machine.  In particular, we
wanted a local copy of [Freebase](https://www.freebase.com/) and
similar knowledge bases.

This project is called "<em>Tiny</em>graph" because the codebase is
tiny.  It just doesn't do much, but it's pretty efficient and easy to
use.

Status: Experimental.

What it can do:

1. Store triples (actually "quads") indexed for (prefixes of)
   subject-property-object, object-property-subject, and
   property-subject-object access.
2. Find edges based on those indexes.
3. Give you a barely functional Go API.
4. Give you a barely functional Javascript-from-HTTP API based on that
   Go API.

That's about it.  The core code is fewer than 1,000 lines of Go.

What it can't do:

1. Query planning.  You just traverse paths.  (But it'll be easy to expose RocksDB's `GetApproximateSizes`.)
2. Fancy [graph algorithms](http://en.wikipedia.org/wiki/Category:Graph_algorithms).
3. Provide much safety.

For comparison, see:

1. [Cayley](https://github.com/google/cayley)
2. [Accumulo](https://accumulo.apache.org/)
3. [Neo4j](http://neo4j.com/)

and many other [graph databases](http://en.wikipedia.org/wiki/Graph_database).

How it works:

1. Uses [RocksDB](http://rocksdb.org/).  Could use any key-value store
   that provides prefix ordering.
2. Does not intern strings.  Instead, uses RocksDB's Snappy
   compression in an simple attempt to mitigate the cost of duplicates
   (while avoiding read+update during each write).
3. Uses the nifty
   [`robertkrimen/otto`](https://github.com/robertkrimen/otto)
   Javascript implementation in Go.


Getting started:

Get [Go](https://golang.org/).

Then try to get [RocksDB](http://rocksdb.org/) built:

```Shell
gcc -v # 4.7+
sudo apt-get install -y libgflags-dev libsnappy-dev zlib1g-dev libbz2-dev
git clone https://github.com/facebook/rocksdb.git
cd rocksdb
make shared_lib
# Suit yourself re installation.
sudo cp librocksdb.so /usr/lib
(cd /usr/include && sudo ln -s ~/rocksdb/include/rocksdb rocksdb)
```

Then

```Shell
go get && go build
```

Quick example:

It's easy to load [WordNet RDF](http://wordnet-rdf.princeton.edu/).
Here's an example of loading a useful subset of WordNet quickly.

```Shell
wget http://wordnet-rdf.princeton.edu/wn31.nt.gz
zcat wn31.nt.gz | grep 'hyponym\|hypernym\|meronym\|holonym\|#label' | gzip > somewn.nt.gz
rm -rf wordnet.db
./tinygraph -config config.wordnet -lang eng -load somewn.nt.gz -serve
```

Then:

```Shell
cat <<EOF > holo.js
function holonyms(term) {
  var label = G.Bs("http://www.w3.org/2000/01/rdf-schema#label");
  var holo = G.Bs("http://wordnet-rdf.princeton.edu/ontology#part_holonym");
  var paths = G.In(label).Out(holo).Out(label).Walk(G.Graph(), G.Vertex(term)).Collect();
  var uniq = {};
  var acc = [];
  for (var i=0; i<paths.length; i++) {
	  var h = paths[i][2].Strings()[2];
	  if (!uniq[h]) {
          uniq[h] = true;
		  acc.push(h);
	  }
  }
  return acc;
}
EOF
curl --data-urlencode 'js@holo.js' http://localhost:8080/js
# Use that stored procedure.
curl --data-urlencode 'js=holonyms("Africa")' http://localhost:8080/js
```

You'll hopefully see

```Javascript
[
    "Barbary",
    "Nubia",
    "Mahgrib",
    "African nation",
    "East Africa",
    "South West Africa",
    "Republic of Angola",
    "Republic of Burundi",
    "Republic of Cameroon",
    "Central African Republic",
    "Tchad",
    "Republic of the Congo",
    "Zaire",
	...
]
```

Check out [`examples/`](examples) for more.  Also see `freebase.sh`.


ToDo:

1. More logging.
2. More test cases.
3. More docs (especially configuration).
4. Buffered Stepper channels.
5. Expose `GetApproximateSizes`.
6. Reorganize for embedded use.
7. Deal with concurrent requests and Javascript.


