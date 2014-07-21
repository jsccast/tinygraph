package main

import (
	"bufio"
	"fmt"
	"github.com/robertkrimen/otto"
	"os"
)

func getString(x interface{}) (string, error) {
	switch vv := x.(type) {
	case string:
		return vv, nil
	default:
		return "", fmt.Errorf("Value %v (%T) not a string", x, x)
	}
}

func repl() {
	scanner := bufio.NewScanner(os.Stdin)
	vm := otto.New()
	og, _ := vm.Object("g = {}")
	og.Set("Triple", func(call otto.FunctionCall) otto.Value {
		x, err := call.Argument(0).Export()
		if err != nil {
			panic(err)
		}
		fmt.Printf("x %T %v\n", x, x)
		switch vv := x.(type) {
		case []interface{}:
			var s string
			var p string
			var o string
			var v string
			if 0 < len(vv) {
				if s, err = getString(vv[0]); err != nil {
					panic(err)
				}
			}
			if 1 < len(vv) {
				if p, err = getString(vv[1]); err != nil {
					panic(err)
				}
			}
			if 2 < len(vv) {
				if o, err = getString(vv[2]); err != nil {
					panic(err)
				}
			}
			if 3 < len(vv) {
				if o, err = getString(vv[3]); err != nil {
					panic(err)
				}
			}

			y, err := vm.ToValue(&Triple{[]byte(s), []byte(p), []byte(o), []byte(v)})
			if err != nil {
				panic(err)
			}
			return y
		}

		return call.Argument(0)
	})

	vm.Set("twoPlus", func(call otto.FunctionCall) otto.Value {
		right, _ := call.Argument(0).ToInteger()
		result, _ := vm.ToValue(2 + right)
		return result
	})
	for scanner.Scan() {
		line := scanner.Text()
		x, err := vm.Run(line)
		fmt.Printf("%v (%v)\n", x, err)
	}
}
