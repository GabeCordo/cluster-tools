package cli

import (
	"fmt"
	"time"
)

const (
	ASCII_BORDER       = ""
	WHITE_SPACE        = " "
	TRAILER            = "| "
	MAX_PRINT_BOX_SIZE = 42
)

func ascii_border(maxPrintBoxSize int) string {
	var s string
	for i := 0; i < maxPrintBoxSize; i++ {
		s += "-"
	}
	return s + "\n"
}

func ascii_column(largestKeyLen int, usedSpace int) string {
	var s string
	for i := 0; i < (largestKeyLen - usedSpace); i++ {
		s += WHITE_SPACE
	}
	return s + TRAILER
}

func ascii_space_ending(usedSpace int, data string, maxPrintBoxSize int) string {
	var s string
	for i := 0; i < (maxPrintBoxSize - (len(data) + usedSpace)); i++ {
		s += WHITE_SPACE
	}
	return s + TRAILER
}

func box(data map[string]string, max ...int) {
	maxBoxSize := MAX_PRINT_BOX_SIZE
	if len(max) == 1 {
		maxBoxSize = max[0]
	}

	largestKey := 0
	for key, _ := range data {
		l := len(key)
		if largestKey < l {
			largestKey = l
		}
	}

	print(ascii_border(maxBoxSize))
	for key, value := range data {
		fmt.Printf("| %s %s%s%s\n", key, ascii_column(largestKey, len(key)), value, ascii_space_ending(largestKey+6, value, maxBoxSize))
		print(ascii_border(maxBoxSize))
	}
}

func carrige() {
	for i := 0; i < 10; i++ {
		fmt.Printf("completion %d \r", i)
		time.Sleep(1 * time.Second)
	}
}
