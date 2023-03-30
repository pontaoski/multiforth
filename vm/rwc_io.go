package vm

import "io"

type ReadWriteCloserIODevice struct {
	Items map[Cell]io.ReadWriteCloser
}

func NewRWCIODevice() *ReadWriteCloserIODevice {
	return &ReadWriteCloserIODevice{map[Cell]io.ReadWriteCloser{}}
}

func (rwc *ReadWriteCloserIODevice) Add(f io.ReadWriteCloser) Cell {
	for num := 0; num < (2 << 15); num++ {
		if _, ok := rwc.Items[Cell(num)]; ok {
			continue
		}
		rwc.Items[Cell(num)] = f
		return Cell(num)
	}
	panic("failed to add readwritecloser")
}

func (rwc *ReadWriteCloserIODevice) Handle(c *Core) {
	command := c.popD()
	switch command {
	case 0: // read byte
		target := c.popD()
		item := rwc.Items[target]

		var b [1]byte
		_, err := item.Read(b[:])
		if err != nil {
			panic(err)
		}
		c.pushD(Cell(b[0]))
	case 1: // write byte
		bit := c.popD()
		target := c.popD()
		item := rwc.Items[target]

		_, err := item.Write([]byte{byte(bit)})
		if err != nil {
			panic(err)
		}
	case 2: // close
		target := c.popD()
		item := rwc.Items[target]
		item.Close()
		delete(rwc.Items, target)
	}
}
