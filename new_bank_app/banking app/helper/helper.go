package helper

import (
	"fmt"
	"strings"
)

type Authorizer interface {
	IsAdminUser() bool
	IsActiveUser() bool
}

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

func NameTrim(s string) string {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return ""
	}
	return trimmed
}

func SafeExec(label string, fn func()) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Recovered in [%s]: %v\n", label, r)
		}
	}()
	fn()
}

var idCounter = 1000

func GenerateUniqueID() int {
	idCounter++
	return idCounter
}
