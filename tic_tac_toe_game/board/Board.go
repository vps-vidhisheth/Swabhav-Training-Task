package board

import (
	"tic_tac_toe/cell"
	"tic_tac_toe/errorhandler"
)

type Board struct {
	Cells [9]*cell.Cell
}

func NewBoard() *Board {
	b := &Board{}
	for i := range b.Cells {
		b.Cells[i] = cell.NewCell()
	}
	return b
}

func (b *Board) IsCellEmpty(index int) (bool, error) {
	if index < 0 || index >= len(b.Cells) {
		return false, errorhandler.NewOutOfBoundsError()
	}
	return b.Cells[index].IsEmpty(), nil
}

func (b *Board) SetCell(index int, symbol string) error {
	if index < 0 || index >= len(b.Cells) {
		return errorhandler.NewOutOfBoundsError()
	}
	b.Cells[index].SetMark(symbol)
	return nil
}

func (b *Board) IsFull() bool {
	for _, c := range b.Cells {
		if c.IsEmpty() {
			return false
		}
	}
	return true
}

func (b *Board) Reset() {
	for _, c := range b.Cells {
		c.Clear()
	}
}

func (b *Board) CheckWin(symbol string) bool {
	winPatterns := [8][3]int{
		{0, 1, 2}, {3, 4, 5}, {6, 7, 8},
		{0, 3, 6}, {1, 4, 7}, {2, 5, 8},
		{0, 4, 8}, {2, 4, 6},
	}
	for _, pattern := range winPatterns {
		if b.Cells[pattern[0]].Mark == symbol &&
			b.Cells[pattern[1]].Mark == symbol &&
			b.Cells[pattern[2]].Mark == symbol {
			return true
		}
	}
	return false
}
