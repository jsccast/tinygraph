// Copyright 2014 The Cayley Authors. All rights reserved.
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

package main

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

func isWhitespace(s uint8) bool {
	return (s == '\t' || s == '\r' || s == ' ')
}
func Parse(str string) *Triple {
	// Skip leading whitespace.
	str = skipWhitespace(str)
	// Check for a comment
	if str != "" && str[0] == '#' {
		return nil
	}
	sub, remainder := getTripleComponent(str)
	if sub == nil {
		return nil
	}
	str = skipWhitespace(remainder)
	pred, remainder := getTripleComponent(str)
	if pred == nil {
		return nil
	}
	str = skipWhitespace(remainder)
	obj, remainder := getTripleComponent(str)
	if obj == nil {
		return nil
	}
	str = skipWhitespace(remainder)
	prov_ptr, remainder := getTripleComponent(str)
	var prov string
	if prov_ptr == nil {
		prov = ""
	} else {
		prov = *prov_ptr
	}
	str = skipWhitespace(remainder)

	// fmt.Printf("line %s %s %s %s\n", *sub, *pred, *obj, prov)

	if str != "" && str[0] == '.' {
		return &Triple{[]byte(*sub), []byte(*pred), []byte(*obj), []byte(prov)}
	}
	return nil
}

func skipWhitespace(str string) string {
	i := 0
	for i < len(str) && isWhitespace(str[i]) {
		i += 1
	}
	return str[i:]
}

func getTripleComponent(str string) (*string, string) {
	if len(str) == 0 {
		return nil, str
	}
	if str[0] == '<' {
		return getUriPart(str[1:])
	} else if str[0] == '"' {
		return getQuotedPart(str[1:])
	} else if str[0] == '.' {
		return nil, str
	} else {
		// Technically not part of the spec. But we do it anyway for convenience.
		return getUnquotedPart(str)
	}
}

func getUriPart(str string) (*string, string) {
	i := 0
	for i < len(str) && str[i] != '>' {
		i += 1
	}
	if i == len(str) {
		return nil, str
	}
	part := str[0:i]
	return &part, str[i+1:]
}

func getQuotedPart(str string) (*string, string) {
	i := 0
	start := 0
	out := ""
	for i < len(str) && str[i] != '"' {
		if str[i] == '\\' {
			out += str[start:i]
			switch str[i+1] {
			case '\\':
				out += "\\"
			case 'r':
				out += "\r"
			case 'n':
				out += "\n"
			case 't':
				out += "\t"
			case '"':
				out += "\""
			default:
				return nil, str
			}
			i += 2
			start = i
			continue
		}
		i += 1
	}
	if i == len(str) {
		return nil, str
	}
	out += str[start:i]
	i += 1
	var remainder string
	if strings.HasPrefix(str[i:], "^^<") {
		// Ignore type, for now
		_, remainder = getUriPart(str[i+3:])
	} else if strings.HasPrefix(str[i:], "@") {
		_, remainder = getUnquotedPart(str[i+1:])
		lang := getLanguage(str[i:])
		if lang != *onlyLang {
			return nil, ""
		}
		// out = lang + "@" + out
	} else {
		remainder = str[i:]
	}

	return &out, remainder
}

func getLanguage(str string) string {
	i := 1
	lang := ""
	for i < len(str) && !isWhitespace(str[i]) {
		lang += str[i : i+1]
		i++
	}
	return lang
}

func getUnquotedPart(str string) (*string, string) {
	i := 0
	initStr := str
	out := ""
	start := 0
	for i < len(str) && !isWhitespace(str[i]) {
		if str[i] == '"' {
			part, remainder := getQuotedPart(str[i+1:])
			if part == nil {
				return part, initStr
			}
			out += str[start:i]
			str = remainder
			i = 0
			start = 0
			out += *part
		}
		i += 1
	}
	out += str[start:i]
	return &out, str[i:]
}

func ReadNQuadsFromReader(c chan *Triple, reader io.Reader) {
	bf := bufio.NewReader(reader)

	nTriples := 0
	line := ""
	for {
		l, pre, err := bf.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(fmt.Errorf("Something bad happened while reading file: %v", err.Error()))
		}
		line += string(l)
		if pre {
			continue
		}
		triple := Parse(line)
		line = ""
		if triple != nil {
			nTriples++
			c <- triple
		}
	}

	close(c)
}
