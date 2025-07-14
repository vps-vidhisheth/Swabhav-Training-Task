package player

import (
	"strings"
	"tic_tac_toe/errorhandler"
)

type Player struct {
	Name   string
	Symbol string
}

func NewPlayer(name, symbol string) (*Player, error) {
	name = strings.TrimSpace(name)
	symbol = strings.TrimSpace(symbol)

	if name == "" {
		return nil, errorhandler.NewEmptyFieldError("Player name")
	}
	if symbol == "" {
		return nil, errorhandler.NewEmptyFieldError("Player symbol")
	}

	return &Player{
		Name:   name,
		Symbol: symbol,
	}, nil
}
