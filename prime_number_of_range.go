package main

import "fmt"

func isPrime(n int) bool {
	if n <= 1 {
		return false
	}
	for i := 2; i*i <= n; i++ {
		if n%i == 0 {
			return false
		}
	}
	return true
}

func main() {
	var start, end int

	fmt.Print("Enter the start of the range: ")
	fmt.Scanln(&start)

	fmt.Print("Enter the end of the range: ")
	fmt.Scanln(&end)

	fmt.Printf("Prime numbers between %d and %d are:\n", start, end)

	for i := start; i <= end; i++ {
		if isPrime(i) {
			fmt.Print(i, " ")
		}
	}
	fmt.Println()
}
