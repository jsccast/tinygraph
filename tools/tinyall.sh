#!/bin/bash

# Send a bunch of compressed splits containing triples to Tinygraph

set -e

ls x*.gz | xargs --max-args=5 ./tinygroup.sh
