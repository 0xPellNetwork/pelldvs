package utils

import (
	"bufio"
	"os"
)

func GetStdInPassword() (string, bool) {
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		// Input is available in the pipe, read from it
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			return scanner.Text(), true
		}
	}
	return "", false
}
