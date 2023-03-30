package main

import (
	"MultiForth/vm"
	"os"
)

func main() {
	v := vm.NewVM()
	d := vm.NewRWCIODevice()
	v.IO[0] = d

	_ = d.Add(os.Stdin)
	_ = d.Add(os.Stdout)
	_ = d.Add(os.Stderr)

	// loc := v.Add([]Cell{
	// 	// Pack(literal, nop, nop, nop),
	// 	// cin,
	// 	// Pack(literal, nop, nop, nop),
	// 	// 0,
	// 	// Pack(literal, nop, nop, nop),
	// 	// 0,
	// 	Pack(literal, nop, nop, nop),
	// 	cout,
	// 	Pack(literal, nop, nop, nop),
	// 	'U',
	// 	Pack(literal, nop, nop, nop),
	// 	1,
	// 	Pack(literal, nop, nop, nop),
	// 	0,
	// 	Pack(doIO, literal, nop, nop),
	// 	cout,
	// 	Pack(literal, nop, nop, nop),
	// 	'\n',
	// 	Pack(literal, nop, nop, nop),
	// 	1,
	// 	Pack(literal, nop, nop, nop),
	// 	0,
	// 	Pack(doIO, die, nop, nop),
	// })

	//v.Spawn(loc)
	<-v.AllDeath
	println("All cores died :(")
}
