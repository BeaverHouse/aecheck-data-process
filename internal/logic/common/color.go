package common

import (
	"fmt"
)

// ANSI color codes
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
)

// PrintSuccess prints a success message in green
func PrintSuccess(text string) {
	fmt.Printf("\033[32m%s\033[0m\n", text)
}

// PrintError prints an error message in red
func PrintError(text string) {
	fmt.Printf("\033[31m%s\033[0m\n", text)
}

// PrintWarning prints a warning message in yellow
func PrintWarning(text string) {
	fmt.Printf("\033[33m%s\033[0m\n", text)
}
