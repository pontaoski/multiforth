package vm

const (
	Nop byte = iota
	Literal
	Dup
	Drop
	Swap
	Push
	Pop
	Jump
	CondJump
	Call
	CondCall
	Ret
	RetIfZero
	Equal
	NotEqual
	LessThan
	GreaterThan
	Fetch
	Store
	Add
	Subtract
	Multiply
	DivideRemainder
	And
	Or
	Xor
	ShiftLeft
	ShiftRight
	Alloc
	ResizeSegment
	Free
	Spawn
	Send
	Recv
	ReadRegister
	WriteRegister
	DoIO
	Die
	Compare
	Copy

	dbg     = 0xFF - 1
	invalid = 0xFF
)
