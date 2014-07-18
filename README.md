A tiny graph database

Well, the code is tiny, but the data can be single-machine big.

Currently requires [`rocksdb`](http://rocksdb.org/), but many other
back ends (e.g., [`levigo`](https://github.com/jmhodges/levigo) or
[`goleveldb`](https://github.com/syndtr/goleveldb)) would be easy to
support.

Also see [Cayley](https://github.com/google/cayley).  I have a
[fork](http://github.csv.comcast.com/jsteph206/cayley) that supports
[`rocksdb`](http://rocksdb.org/).

Currently I'm using this code to try to load all of
[Freebase](https://developers.google.com/freebase/data).

