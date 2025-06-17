package main

import (
	"fmt"
	"time"
)

func main() {
	currentTime := time.Now()
	hour := currentTime.Hour()
	minute := currentTime.Minute()
	second := currentTime.Second()

	fmt.Printf("Current Time: %02d:%02d:%02d\n", hour, minute, second)

	if (hour > 6 && hour < 11) || (hour == 6 && (minute > 0 || second > 0)) || (hour == 11 && minute == 0 && second == 0) {
		fmt.Println("Good morning!")
	} else if (hour > 11 && hour < 16) || (hour == 11 && (minute > 0 || second > 0)) || (hour == 16 && minute == 0 && second == 0) {
		fmt.Println("Good afternoon!")
	} else if (hour > 16 && hour < 21) || (hour == 16 && (minute > 0 || second > 0)) || (hour == 21 && minute == 0 && second == 0) {
		fmt.Println("Good evening!")
	} else {
		fmt.Println("Good night!")
	}
}
