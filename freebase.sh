#!/bin/bash

# Get a Freebase dump from https://developers.google.com/freebase/data
DUMP=freebase-rdf-2014-07-13-00-00.gz
wget http://commondatastorage.googleapis.com/freebase-public/rdf/$DUMP

# Process it.
rm -rf test.db
mkdir -p par && (cd par && rm -f *.par)
zcat $DUMP | \
  parallel --jobs 6 --pipe --files --block 100M --tmpdir par gzip | \
  xargs -n 6 echo | \
  unbuffer -p sed 's/ /,/g' | \
  xargs -n 1 -I FILES bash -c '../tinygraph -gzip -config config.freebase -silent-ignore -load FILES; rm `echo FILES | tr -d "[:space:]" | sed "s/,/ /g"`'

