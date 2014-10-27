package tinygraph

import "flag"

var onlyLang = flag.String("lang", "eng", "Only get these strings ('en' for Freebase; 'eng' for WordNet)")
var gzipin = flag.Bool("gzip", false, "Input triple files are gzipped")
var ignoreSilently = flag.Bool("silent-ignore", true, "Don't report when ingoring a triple")
var chanBufferSize = flag.Int("chanbuf", 16, "Traversal emission buffer")
