#!/bin/bash

# Uncompress given files to `/dev/shm`, and then run Tinygraph.
# Assumes `config.json` `triple_files` point to the right files.

set -e

DIR=/dev/shm/rocks/in
mkdir -p $DIR
rm -f $DIR/*

echo tinygroup "$@"

time parallel -I F "zcat F > $DIR/F" ::: "$@"

i=0
for F in `ls $DIR`; do
   i=$((i+1))
   mv $DIR/$F $DIR/$i
done

time ./tinygraph 2>&1 | tee -a tg.log
