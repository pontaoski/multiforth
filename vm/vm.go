package vm

import (
	"errors"
	"fmt"
	"runtime/debug"
)

type Cell uint64
type HalfCell uint32

var no Cell = 0
var yes Cell = ^(Cell(0))
var errOutOfIndex error = errors.New("out of index")
var errDie error = errors.New("die instruction ran")

type VM struct {
	Memory   map[HalfCell][]Cell
	Cores    map[HalfCell]*Core
	AllDeath chan struct{}
	IO       map[Cell]IODevice
}

func NewVM() *VM {
	return &VM{map[HalfCell][]Cell{}, map[HalfCell]*Core{}, make(chan struct{}), map[Cell]IODevice{}}
}

func dumpAddr(a Cell) string {
	num := (a >> 32) & 0xffffffff
	idx := a & 0xffffffff
	return fmt.Sprintf("%d::%d", num, idx)
}

func Pack(a, b, c, d byte) Cell {
	ac, bc, cc, dc := Cell(a), Cell(b), Cell(c), Cell(d)

	return ac | bc<<8 | cc<<16 | dc<<24
}

func (v *VM) Unregister(c *Core) {
	delete(v.Cores, HalfCell(c.ID))
	if len(v.Cores) == 0 {
		v.AllDeath <- struct{}{}
	}
}

func (v *VM) Spawn(at Cell) (Cell, *Core) {
	for num := 0; num < (2 << 15); num++ {
		if _, ok := v.Cores[HalfCell(num)]; ok {
			continue
		}
		c := &Core{}
		c.ID = Cell(num)
		c.From = v
		c.InstructionPointer = at
		c.Mailbox = make(chan Cell)
		c.Run()
		v.Cores[HalfCell(num)] = c
		return Cell(num), c
	}
	panic("failed to alloc core")
}

func (v *VM) Add(cs []Cell) Cell {
	a, b := v.Alloc(Cell(len(cs)))
	copy(b, cs)
	return a
}

func (v *VM) ResizeSegment(addr, size Cell) {
	segment := (addr >> 32) & 0xffffffff
	mem := v.Memory[HalfCell(segment)]
	nmem := make([]Cell, size)
	copy(nmem, mem)
}

func (v *VM) Alloc(s Cell) (Cell, []Cell) {
	for num := 0; num < (2 << 15); num++ {
		if _, ok := v.Memory[HalfCell(num)]; ok {
			continue
		}
		v.Memory[HalfCell(num)] = make([]Cell, s)
		return (Cell(num) << 32), v.Memory[HalfCell(num)]
	}
	panic("failed to alloc")
}

func (v *VM) Free(a Cell) {
	num := (a >> 32) & 0xffffffff
	delete(v.Memory, HalfCell(num))
}

func (vm *VM) Write(a Cell, v Cell) {
	num := (a >> 32) & 0xffffffff
	idx := a & 0xffffffff
	arr := vm.Memory[HalfCell(num)]
	if int(idx) >= len(arr) {
		panic(errOutOfIndex)
	}
	arr[idx] = v
}

func (v *VM) Read(c Cell) Cell {
	num := (c >> 32) & 0xffffffff
	idx := c & 0xffffffff
	arr := v.Memory[HalfCell(num)]
	if int(idx) >= len(arr) {
		panic(errOutOfIndex)
	}
	return arr[idx]
}

type Core struct {
	From *VM
	ID   Cell

	Registers          [24]Cell
	Data               [32]Cell
	DataPointer        Cell
	Address            [256]Cell
	AddressPointer     Cell
	InstructionPointer Cell
	Mailbox            chan Cell
}

func cond(b bool) Cell {
	if b {
		return yes
	}
	return no
}

func pushC(a []Cell, p *Cell, v Cell) {
	*p++
	a[*p] = v
}

func popC(a []Cell, p *Cell) Cell {
	r := a[*p]
	*p--
	return r
}

func (c *Core) pushD(v Cell) {
	pushC(c.Data[:], &c.DataPointer, v)
}

func (c *Core) popD() Cell {
	return popC(c.Data[:], &c.DataPointer)
}

func (c *Core) pushA(v Cell) {
	pushC(c.Address[:], &c.AddressPointer, v)
}

func (c *Core) popA() Cell {
	return popC(c.Address[:], &c.AddressPointer)
}

func (c *Core) ProcessInstruction(instruction byte) {
	switch instruction {
	case Nop:
	case Literal:
		c.InstructionPointer++
		c.pushD(c.From.Read(c.InstructionPointer))
	case Dup:
		m := c.popD()
		c.pushD(m)
		c.pushD(m)
	case Drop:
		c.popD()
	case Swap:
		a, b := c.popD(), c.popD()
		c.pushD(a)
		c.pushD(b)
	case Push:
		c.pushA(c.popD())
	case Pop:
		c.pushD(c.popA())
	case Jump:
		c.InstructionPointer = c.popD() - 1
	case CondJump:
		addr := c.popD()
		cond := c.popD()
		if cond != 0 {
			c.InstructionPointer = addr - 1
		}
	case Call:
		c.pushA(c.InstructionPointer)
		c.InstructionPointer = c.popD() - 1
	case CondCall:
		addr := c.popD()
		cond := c.popD()
		if cond != 0 {
			c.pushA(c.InstructionPointer)
			c.InstructionPointer = addr - 1
		}
	case Ret:
		c.InstructionPointer = c.popA()
	case RetIfZero:
		if c.Data[c.DataPointer] == 0 {
			c.popD()
			c.InstructionPointer = c.popA()
		}
	case Equal:
		a := c.popD()
		b := c.popD()
		c.pushD(cond(b == a))
	case NotEqual:
		a := c.popD()
		b := c.popD()
		c.pushD(cond(b != a))
	case LessThan:
		a := c.popD()
		b := c.popD()
		c.pushD(cond(b < a))
	case GreaterThan:
		a := c.popD()
		b := c.popD()
		c.pushD(cond(b > a))
	case Fetch:
		c.pushD(c.From.Read(c.popD()))
	case Store:
		a := c.popD()
		v := c.popD()
		c.From.Write(a, v)
	case Add:
		a := c.popD()
		b := c.popD()
		c.pushD(b + a)
	case Subtract:
		a := c.popD()
		b := c.popD()
		c.pushD(b - a)
	case Multiply:
		a := c.popD()
		b := c.popD()
		c.pushD(b * a)
	case DivideRemainder:
		a := c.popD()
		b := c.popD()
		if a == 0 {
			c.pushD(no)
			c.pushD(no)
		} else {
			c.pushD(b % a)
			c.pushD(b / a)
		}
	case And:
		a := c.popD()
		b := c.popD()
		c.pushD(b & a)
	case Or:
		a := c.popD()
		b := c.popD()
		c.pushD(b | a)
	case Xor:
		a := c.popD()
		b := c.popD()
		c.pushD(b ^ a)
	case ShiftLeft:
		a := c.popD()
		b := c.popD()
		c.pushD(b << a)
	case ShiftRight:
		a := c.popD()
		b := c.popD()
		c.pushD(b >> a)
	case Alloc:
		s := c.popD()
		addr, _ := c.From.Alloc(s)
		c.pushD(addr)
	case ResizeSegment:
		size := c.popD()
		addr := c.popD()
		c.From.ResizeSegment(addr, size)
		c.pushD(addr)
	case Free:
		a := c.popD()
		c.From.Free(a)
	case Spawn:
		ip := c.popD()
		idx, _ := c.From.Spawn(ip)
		c.pushD(idx)
	case Send:
		target := c.popD()
		msg := c.popD()
		c.From.Cores[HalfCell(target)].Mailbox <- msg
	case Recv:
		c.pushD(<-c.Mailbox)
	case dbg:
		c.Debug()
	case ReadRegister:
		c.pushD(c.Registers[c.popD()])
	case WriteRegister:
		a := c.popD()
		b := c.popD()
		c.Registers[a] = b
	case DoIO:
		dev := c.popD()
		c.From.IO[dev].Handle(c)
	case Die:
		panic(errDie)
	case Compare:
		l := c.popD()
		dest := c.popD()
		src := c.popD()

		for i := Cell(0); i < l; i++ {
			if c.From.Read(dest+l) != c.From.Read(src+l) {
				c.pushD(no)
			}
		}
		c.pushD(yes)
	case Copy:
		l := c.popD()
		dest := c.popD()
		src := c.popD()

		for i := Cell(0); i < l; i++ {
			c.From.Write(src+i, c.From.Read(dest+i))
		}
	default:
		panic("bad instruction")
	}
}

func (c *Core) Debug() {
	println("Core debug")
	print("Registers: ")
	for _, reg := range c.Registers {
		print(reg, " ")
	}
	println()
	print("Data stack: ")
	for i := 1; i <= int(c.DataPointer); i++ {
		print(c.Data[i], " ")
	}
	println()
	print("Address stack: ")
	for i := 1; i <= int(c.AddressPointer); i++ {
		print(c.Address[i], " ")
	}
	println()
	println("Instruction pointer:", dumpAddr(c.InstructionPointer))
}

func (c *Core) ProcessBundle(bundle Cell) {
	c.ProcessInstruction((byte)((bundle >> 0) & 0xff))
	c.ProcessInstruction((byte)((bundle >> 8) & 0xff))
	c.ProcessInstruction((byte)((bundle >> 16) & 0xff))
	c.ProcessInstruction((byte)((bundle >> 24) & 0xff))
}

func (c *Core) Run() {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				if r == errOutOfIndex {
					fmt.Printf("Core %d crashed: out of bounds access\n", c.ID)
					c.Debug()
				} else if r == errDie {
					// nothing
				} else {
					fmt.Printf("Core %d crashed\n", c.ID)
					fmt.Println(r)
					println(string(debug.Stack()))
				}
			}
			c.From.Unregister(c)
		}()
		for {
			bundle := c.From.Read(c.InstructionPointer)
			c.ProcessBundle(bundle)
			c.InstructionPointer++
		}
	}()
}
