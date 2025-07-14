package game

import (
	"fmt"
	"tic_tac_toe/board"
	"tic_tac_toe/errorhandler"
	"tic_tac_toe/player"
)

type Game struct {
	Players    [2]*player.Player
	Board      *board.Board
	Turn       int
	Winner     *player.Player
	IsDraw     bool
	IsGameOver bool
}

func recoverFromPanic(context string) {
	if r := recover(); r != nil {
		fmt.Printf("Recovered from panic in %s: %v\n", context, r)
	}
}

func CreateGame(name1, symbol1, name2, symbol2 string) (g *Game, err error) {
	defer recoverFromPanic("CreateGame")

	p1, err := player.NewPlayer(name1, symbol1)
	if err != nil {
		return nil, err
	}
	p2, err := player.NewPlayer(name2, symbol2)
	if err != nil {
		return nil, err
	}
	if p1.Symbol == p2.Symbol {
		return nil, errorhandler.NewSameSymbolError()
	}

	return &Game{
		Players:    [2]*player.Player{p1, p2},
		Board:      board.NewBoard(),
		Turn:       0,
		Winner:     nil,
		IsDraw:     false,
		IsGameOver: false,
	}, nil
}

func (g *Game) CurrentPlayer() *player.Player {
	return g.Players[g.Turn%2]
}

func (g *Game) PlayTurn(pos int) error {
	if g.IsGameOver {
		return errorhandler.NewGameOverError()
	}

	player := g.CurrentPlayer()
	g.Board.SetCell(pos, player.Symbol)

	if g.checkWin(player.Symbol) {
		g.Winner = player
		g.IsGameOver = true
	} else if g.Board.IsFull() {
		g.IsDraw = true
		g.IsGameOver = true
	} else {
		g.Turn++
	}

	return nil
}

func (g *Game) checkWin(symbol string) bool {
	return g.Board.CheckWin(symbol)
}

func (g *Game) PrintBoard() {
	defer recoverFromPanic("PrintBoard")

	fmt.Println()
	for i, cell := range g.Board.Cells {
		mark := cell.Mark
		if mark == "" {
			mark = " "
		}
		fmt.Printf(" %s ", mark)
		if i%3 != 2 {
			fmt.Print("|")
		}
		if i%3 == 2 && i != 8 {
			fmt.Println("\n-----------")
		}
	}
	fmt.Println()
}

func (g *Game) Reset() {
	defer recoverFromPanic("Reset")

	g.Board.Reset()
	g.Turn = 0
	g.Winner = nil
	g.IsDraw = false
	g.IsGameOver = false
	fmt.Println("\nGame has been reset.")
}

func (g *Game) PrintStatus() {
	defer recoverFromPanic("PrintStatus")

	if g.IsGameOver {
		if g.IsDraw {
			fmt.Println("Game ended in a draw.")
		} else {
			fmt.Printf("Player %s (%s) wins!\n", g.Winner.Name, g.Winner.Symbol)
		}
	} else {
		fmt.Println("Game is ongoing.")
	}
}

func PlayMove(g *Game, pos int) {
	defer recoverFromPanic("PlayMove")

	err := g.PlayTurn(pos)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	g.PrintBoard()
	g.PrintStatus()
	fmt.Println()
}
