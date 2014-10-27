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

import "flag"

var onlyLang = flag.String("lang", "eng", "Only get these strings ('en' for Freebase; 'eng' for WordNet)")
var gzipin = flag.Bool("gzip", false, "Input triple files are gzipped")
var ignoreSilently = flag.Bool("silent-ignore", true, "Don't report when ingoring a triple")
var chanBufferSize = flag.Int("chanbuf", 16, "Traversal emission buffer")
