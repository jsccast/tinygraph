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

// A fairly fragile RDF triple (quad) parser.

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"strings"
)

func ParseTriple(s string) (*Triple, error) {
	more := strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}
	if s[0] == '#' {
		return nil, nil
	}

	sub, more, err := parseTriplePart(more)
	if err != nil {
		log.Printf("ParseTriple error %v on '%s' at sub '%s'", err, s, more)
		return nil, err
	}
	if sub == "" {
		if !*ignoreSilently {
			log.Printf("ParseTriple ignoring '%s' due to sub", s)
		}
		return nil, nil
	}
	// log.Println(sub + "; " + more)

	pred, more, err := parseTriplePart(more)
	if err != nil {
		log.Printf("ParseTriple error %v on '%s' at pred '%s'", err, s, more)
		return nil, err
	}
	if pred == "" {
		if !*ignoreSilently {
			log.Printf("ParseTriple ignoring '%s' due to pred", s)
		}
		return nil, nil
	}
	// log.Println(pred + "; " + more)

	obj, more, err := parseTriplePart(more)
	if err != nil {
		log.Printf("ParseTriple error %v on '%s' at obj '%s'", err, s, more)
		return nil, err
	}
	if obj == "" {
		if !*ignoreSilently {
			log.Printf("ParseTriple ignoring '%s' due to obj", s)
		}
		return nil, nil
	}
	// log.Println(obj + "; " + more)

	meta, more, err := parseTriplePart(more)
	if err != nil {
		log.Printf("ParseTriple error %v on '%s' at meta '%s'", err, s, more)
		return nil, err
	}
	// log.Println(meta + "; " + more)

	if more != "" && more != "." {
		return nil, fmt.Errorf("Triple '%s' not terminated properly with '%s'",
			s, more)
	}

	t := Triple{[]byte(sub), []byte(pred), []byte(obj), []byte(meta)}
	return &t, nil
}

func parseTriplePart(s string) (string, string, error) {
	if len(s) == 0 {
		return "", s, fmt.Errorf("Empty triple part")
	}
	s = strings.TrimSpace(s)
	switch s[0] {
	case '<':
		return parseURI(s[1:])
	case '"':
		return parseQuoted(s[1:])
	case '.':
		return "", s, nil
	default:
		return parseUnquoted(s)
	}
}

func parseTerminated(s, terminator string) (string, string, error) {
	end := strings.Index(s, terminator)
	if end < 0 {
		return "", s, fmt.Errorf("Unterminated part '%s'", terminator+s)
	}
	return s[0:end], s[end+1:], nil
}

func parseURI(s string) (string, string, error) {
	return parseTerminated(s, ">")
}

func parseQuoted(s string) (string, string, error) {
	i := 0
	n := len(s)
	raw := make([]byte, 0, len(s))
LOOP:
	for i < n {
		switch s[i] {
		case '\\':
			if i+1 == n {
				return "", s, fmt.Errorf("Missing escape at %d in '%s'", s)
			}
			switch s[i+1] {
			case '"':
				raw = append(raw, '"')
			case 'r':
				raw = append(raw, '\r')
			case 'n':
				raw = append(raw, '\n')
			case 't':
				raw = append(raw, '\t')
			case '\\':
				raw = append(raw, '\\')
			case 'u':
				raw = append(raw, '?') // ToDo: parse
			default:
				return "", s, fmt.Errorf("Bad escape '%c' at %d in '%s'", s[i], i, s)
			}
			i++
		case '"':
			break LOOP
		default:
			raw = append(raw, s[i])
		}
		i++
	}

	quoted := s[0:i]
	more := s[i+1:]

	if strings.HasPrefix(more, "^^<") {
		// Ignoring type (given by URI).
		var err error
		_, more, err = parseURI(more[3:])
		if err != nil {
			return "", s, err
		}
	} else if strings.HasPrefix(more, "@") {
		if len(more) == 1 {
			return "", s, fmt.Errorf("Missing language after @ in '%s'", s)
		}
		end := strings.IndexAny(more, "\t .")
		if end < 0 {
			end = len(more)
		}
		lang := more[1:end]
		more = more[end:]

		if lang != *onlyLang {
			return "", "", nil
		}
	}

	return quoted, more, nil
}

func parseUnquoted(s string) (string, string, error) {
	end := strings.IndexAny(s, "\t\r ")
	if end < 0 {
		return "", "", fmt.Errorf("Unterminated token '%s'", s)
	}
	return s[0:end], s[end:], nil
}

func ParseTriples(c chan *Triple, reader io.Reader) error {
	r := bufio.NewReader(reader)
	defer close(c)

	for {
		line, err := r.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		triple, err := ParseTriple(strings.TrimSuffix(line, "\n"))
		if err != nil {
			log.Printf("ParseTriples error %v on '%s'", err, line)
			continue
		}
		if triple == nil {
			continue
		}

		c <- triple
	}

	return nil
}
