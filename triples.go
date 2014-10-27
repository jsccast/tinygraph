package tinygraph

import (
	"fmt"
)

type Triple struct {
	S []byte
	P []byte
	O []byte
	V []byte
}

func (t *Triple) Copy() *Triple {
	return &Triple{t.S, t.P, t.O, t.V}
}

func TripleFromStrings(args ...string) *Triple {
	var s []byte
	var p []byte
	var o []byte
	var v []byte
	n := len(args)

	if 0 < n {
		s = []byte(args[0])
	}
	if 1 < n {
		p = []byte(args[1])
	}
	if 2 < n {
		o = []byte(args[2])
	}
	if 3 < n {
		v = []byte(args[3])
	}

	return &Triple{s, p, o, v}
}

// Not func(t *Triple) so Otto can find this method.
func (t Triple) Strings() []string {
	acc := make([]string, 0, 4)
	acc = append(acc, string(t.S))
	acc = append(acc, string(t.P))
	acc = append(acc, string(t.O))
	acc = append(acc, string(t.V))
	return acc
}

// Not func(t *Triple) so Otto can find this method.
func (t Triple) String() string {
	return "<'" +
		string(t.S) + "','" +
		string(t.P) + "','" +
		string(t.O) + "','" +
		string(t.V) + "'>"
}

func TripleFromBytes(k []byte, v []byte) *Triple {

	start := 0
	at := 0

	for 0 < k[at] {
		at++
	}
	s := k[start:at]
	at++
	start = at

	for 0 < k[at] {
		at++
	}
	p := k[start:at]
	at++
	start = at

	for 0 < k[at] {
		at++
	}
	o := k[start:at]
	at++

	return &Triple{s, p, o, v}
}

func (t *Triple) Key() []byte {
	k := make([]byte, 0, len(t.S)+len(t.P)+len(t.O)+3)
	k = append(k, t.S...)
	k = append(k, byte(0))
	k = append(k, t.P...)
	k = append(k, byte(0))
	k = append(k, t.O...)
	k = append(k, byte(0))
	return k
}

func (t *Triple) Val() []byte {
	return t.V
}

func (t *Triple) StartKey() []byte {
	k := make([]byte, 0, len(t.S)+len(t.P)+len(t.O)+3)
	if len(t.S) == 0 {
		return k
	}
	k = append(k, t.S...)
	k = append(k, byte(0))

	k = append(k, t.P...)
	k = append(k, byte(0))

	k = append(k, t.O...)
	k = append(k, byte(0))

	return k
}

func (t *Triple) KeyPrefix() []byte {
	prefix := t.StartKey()
	i := len(prefix) - 1
	for 0 <= i && prefix[i] == 0 {
		i--
	}

	return prefix[0 : i+2]
}

type Index byte

const (
	SPO = iota
	OPS
	PSO
)

// Does not copy!
func (t *Triple) Permute(index Index) *Triple {
	switch index {
	case SPO:
	case OPS:
		s := t.S
		t.S = t.O
		t.O = s
	case PSO:
		s := t.S
		t.S = t.P
		t.P = s
	}

	return t
}

func PrintTriple(t *Triple) bool {
	fmt.Println(t.Strings())
	return true
}
