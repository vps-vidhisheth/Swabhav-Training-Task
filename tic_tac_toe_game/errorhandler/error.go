package errorhandler

import "fmt"

type OutOfBoundsError struct{}

func (e *OutOfBoundsError) Error() string {
	return "Cell index out of bounds"
}

func NewOutOfBoundsError() error {
	return &OutOfBoundsError{}
}

type CellOccupiedError struct{}

func (e *CellOccupiedError) Error() string {
	return "Cell is already occupied"
}

func NewCellOccupiedError() error {
	return &CellOccupiedError{}
}

type GameOverError struct{}

func (e *GameOverError) Error() string {
	return "Game is already over"
}

func NewGameOverError() error {
	return &GameOverError{}
}

type InvalidPositionError struct{}

func (e *InvalidPositionError) Error() string {
	return "Invalid cell position"
}

func NewInvalidPositionError() error {
	return &InvalidPositionError{}
}

type SameSymbolError struct{}

func (e *SameSymbolError) Error() string {
	return "Players must use different symbols"
}

func NewSameSymbolError() error {
	return &SameSymbolError{}
}

type EmptyFieldError struct {
	Field string
}

func (e *EmptyFieldError) Error() string {
	return fmt.Sprintf("%s cannot be empty", e.Field)
}

func NewEmptyFieldError(field string) error {
	return &EmptyFieldError{Field: field}
}
