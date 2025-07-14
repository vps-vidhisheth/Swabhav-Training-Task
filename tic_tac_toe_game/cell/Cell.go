package cell

type Cell struct {
	Mark string
}

func NewCell() *Cell {
	return &Cell{Mark: ""}
}

func (c *Cell) IsEmpty() bool {
	return c.Mark == ""
}

func (c *Cell) SetMark(symbol string) bool {
	if c.IsEmpty() {
		c.Mark = symbol
		return true
	}
	return false
}

func (c *Cell) Clear() {
	c.Mark = ""
}

// surprise , insight , pride , gratitude
// 24 - 40 thousand thoughts a day
