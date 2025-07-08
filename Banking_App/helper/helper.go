package helper

import (
	"fmt"
	"strings"
)

// --- Authorization Interface ---

type Authorizer interface {
	IsAdminUser() bool
	IsActiveUser() bool
}

// --- Pagination Bounds ---

func PaginationBounds(page, pageSize, total int) (start int, end int) {
	if page < 1 {
		page = 1
	}
	start = (page - 1) * pageSize
	end = start + pageSize
	if start >= total {
		start = total
	}
	if end > total {
		end = total
	}
	return
}

// --- Float Comparison ---

func FloatEquals(a, b float64) bool {
	const epsilon = 0.000001
	if a > b {
		return a-b < epsilon
	}
	return b-a < epsilon
}

// --- Name Trimming ---

func NameTrim(s string) string {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return ""
	}
	return trimmed
}

// --- Safe Execution Wrapper ---

func SafeExec(label string, fn func()) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Recovered in [%s]: %v\n", label, r)
		}
	}()
	fn()
}
