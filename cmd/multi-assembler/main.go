package main

import (
	"MultiForth/vm"
	"os"
	"strings"
)

var instructions = map[string]byte{
	"..": vm.Nop,
	"Li": vm.Literal,
	"Du": vm.Dup,
	"Dr": vm.Drop,
	"Sw": vm.Swap,
	"Pu": vm.Push,
	"Po": vm.Pop,
	"Ju": vm.Jump,
	"Co": vm.CondJump,
	"Ca": vm.Call,
	"Cc": vm.CondCall,
	"Re": vm.Ret,
	"Zr": vm.RetIfZero,
	"Eq": vm.Equal,
	"Ne": vm.NotEqual,
	"Le": vm.LessThan,
	"Gr": vm.GreaterThan,
	"Fe": vm.Fetch,
	"St": vm.Store,
	"Ad": vm.Add,
	"Su": vm.Subtract,
	"Mu": vm.Multiply,
	"Di": vm.DivideRemainder,
	"An": vm.And,
	"Or": vm.Or,
	"Xo": vm.Xor,
	"Sl": vm.ShiftLeft,
	"Sr": vm.ShiftRight,
	"Al": vm.Alloc,
	"Rs": vm.ResizeSegment,
	"Fr": vm.Free,
	"Sp": vm.Spawn,
	"Se": vm.Send,
	"Rc": vm.Recv,
	"Rr": vm.ReadRegister,
	"Wr": vm.WriteRegister,
	"Do": vm.DoIO,
	"De": vm.Die,
	"Cm": vm.Compare,
	"Cp": vm.Copy,
}

func main() {
	fi, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	for _, line := range strings.Split(string(fi), "\n") {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		switch line[0] {
		case '#': // comment
		case 'i': // instruction bundle
			insts := strings.Fields(line)[1:]
			if len(insts) != 4 {
				println(len(insts))
				panic("invalid instruction length")
			}
		case ':': // define reference
		case 'r': // use reference
		case 'n': // number
		case 'c': // define constant
		case 'u': // use constant
		}
	}
}
