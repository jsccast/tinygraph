package main

import (
	"fmt"
	"testing"
)

func TestQuads(t *testing.T) {
	triple, err := ParseTriple(`<http://wordnet-rdf.princeton.edu/wn31/100003553-n> <http://wordnet-rdf.princeton.edu/ontology#part_holonym> <http://wordnet-rdf.princeton.edu/wn31/103898588-n> .
`)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(triple.String())
}
