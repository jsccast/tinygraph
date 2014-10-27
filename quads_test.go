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
