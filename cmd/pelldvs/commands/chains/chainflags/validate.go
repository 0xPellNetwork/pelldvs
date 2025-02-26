package chainflags

import (
	"fmt"

	"github.com/spf13/cobra"
)

func RequireFromFlagPersistentForCmds(cmds ...*cobra.Command) error {
	for _, cmd := range cmds {
		cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
			return RequireFromFlag()
		}
	}
	return nil
}

func RequireFromFlag() error {
	if FromKeyNameFlag.Value == "" {
		return fmt.Errorf("required flag `%s` not set", FromKeyNameFlag.Name)
	}
	return nil
}

type IFlagBase interface {
	GetName() string
}

func MarkFlagsAreRequired(cmd *cobra.Command, flags ...IFlagBase) error {
	for _, flag := range flags {
		fname := flag.GetName()
		if fname == "" {
			return fmt.Errorf("flag name is empty")
		}
		_ = cmd.MarkFlagRequired(fname)
	}
	return nil
}
