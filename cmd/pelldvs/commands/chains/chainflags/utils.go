package chainflags

import (
	"os"
)

func getEnvAny(names ...string) string {
	if len(names) == 0 {
		return ""
	}
	for _, name := range names {
		if value := os.Getenv(name); value != "" {
			return value
		}
	}
	return ""
}

func NewAliases(alias ...string) []string {
	return alias
}
