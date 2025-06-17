package main

import (
	"fmt"
	"math"
)

func main() {
	arr := []int{10, 5, 20, 8, 15, 18}
	n := len(arr)

	if n < 2 {
		fmt.Println("Array must contain at least two elements.")
		return
	}

	largest := math.MinInt64
	secondLargest := math.MinInt64

	for i := 0; i < n; i++ {
		if arr[i] > largest {
			secondLargest = largest
			largest = arr[i]
		} else if arr[i] > secondLargest && arr[i] != largest {
			secondLargest = arr[i]
		}
	}

	if secondLargest == math.MinInt64 {
		fmt.Println("There is no second largest number (all elements might be equal).")
	} else {
		fmt.Println("The second largest number is:", secondLargest)
	}
}
