package main

import (
	"fmt"
	"tic_tac_toe/game"
)

func main() {
	fmt.Println("Creating Game 1 ")
	g1, err := game.CreateGame("Vidhi", "X", "Riya", "O")
	if err != nil {
		fmt.Println("Error creating game 1:", err)
		return
	}

	fmt.Println("Game 1 ")
	game.PlayMove(g1, 0)
	game.PlayMove(g1, 4)
	game.PlayMove(g1, 1)

	fmt.Println("Creating Game 2 ")
	g2, err := game.CreateGame("Shruti", "X", "Srushti", "O")
	if err != nil {
		fmt.Println("Error creating game 2:", err)
		return
	}

	fmt.Println(" Game 2 - Playing till draw")
	game.PlayMove(g2, 0)
	game.PlayMove(g2, 1)
	game.PlayMove(g2, 2)
	game.PlayMove(g2, 4)
	game.PlayMove(g2, 3)
	game.PlayMove(g2, 5)
	game.PlayMove(g2, 7)
	game.PlayMove(g2, 6)
	game.PlayMove(g2, 8)

	fmt.Println(" Game 1 - One more move")
	game.PlayMove(g1, 2)

	fmt.Println("Game 1 - Reset and Play Until Player 2 Wins")
	g1.Reset()
	game.PlayMove(g1, 0)
	game.PlayMove(g1, 1)
	game.PlayMove(g1, 3)
	game.PlayMove(g1, 4)
	game.PlayMove(g1, 8)
	game.PlayMove(g1, 7)

	fmt.Println("Final Status of Game 1")
	g1.PrintBoard()
	g1.PrintStatus()
}
