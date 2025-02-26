package utils

import (
	"bufio"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/0xPellNetwork/pelldvs/libs/cli"
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

func GetHomeDir(cmd *cobra.Command) string {
	var home string
	home, _ = cmd.Flags().GetString(cli.HomeFlag)
	if home != "" {
		return home
	}

	home = GetEnvAny("PELLDVS_HOME", "PELLDVSHOME")
	if home != "" {
		return home
	}

	return fmt.Sprintf("%s/.pelldvs", os.ExpandEnv("$HOME"))
}

func GetEnvAny(names ...string) string {
	for _, name := range names {
		if value := os.Getenv(name); value != "" {
			return value
		}
	}
	return ""
}
