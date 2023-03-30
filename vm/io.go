package vm

type IODevice interface {
	Handle(c *Core)
}
