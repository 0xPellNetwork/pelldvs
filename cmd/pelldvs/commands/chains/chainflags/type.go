package chainflags

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

type StringFlag struct {
	Name    string
	Aliases []string
	Usage   string
	Value   string
	Default string
	EnvVars []string
}

func (f *StringFlag) GetName() string {
	return f.Name
}

func (f *StringFlag) AddToCmdFlag(cmds ...*cobra.Command) *StringFlag {
	for _, cmd := range cmds {
		cmd.Flags().StringVar(&f.Value, f.Name, f.Default, f.Usage)
		for _, alias := range f.Aliases {
			cmd.Flags().StringVar(&f.Value, alias, f.Default, fmt.Sprintf("\t[alias for '%s']", f.Name))
			_ = cmd.Flags().MarkHidden(alias)
		}
	}
	return f
}

func (f *StringFlag) AddToCmdPersistentFlags(cmds ...*cobra.Command) *StringFlag {
	for _, cmd := range cmds {
		cmd.PersistentFlags().StringVar(&f.Value, f.Name, f.Default, f.Usage)
		for _, alias := range f.Aliases {
			cmd.PersistentFlags().StringVar(&f.Value, alias, f.Default, fmt.Sprintf("\t[alias for '%s']", f.Name))
			_ = cmd.PersistentFlags().MarkHidden(alias)
		}
	}
	return f
}

func (f *StringFlag) MarkRequired(cmds ...*cobra.Command) error {
	for _, cmd := range cmds {
		err := MarkFlagsAreRequired(cmd, f)
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *StringFlag) GetValue() string {
	// if value is set by flag, return it
	if f.Value != "" {
		return f.Value
	}

	// if value is set by env, return it
	v := getEnvAny(f.EnvVars...)
	if v != "" {
		return v
	}

	// return default value
	return f.Default
}

// SetValue sets the value of the flag, if it is not already set, make sure to call this after all flags are parsed
func (f *StringFlag) SetValue() {
	// if value is set by flag, return it
	if f.Value != "" {
		return
	}

	// if value is set by env, return it
	v := getEnvAny(f.EnvVars...)
	if v != "" {
		f.Value = v
	}

	f.Value = f.Default
}

func (f *StringFlag) GetBool() bool {
	return parseBoolValueFromString(f.GetValue())
}

func parseBoolValueFromString(s string) bool {
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)

	if len(s) == 0 {
		return false
	}

	for _, v := range []string{"true", "t", "yes", "y", "1"} {
		if s == v {
			return true
		}
	}
	return false
}

type IntFlag struct {
	Name    string
	Aliases []string
	Usage   string
	Value   int
	Default int
	EnvVars []string
}

func (f *IntFlag) GetName() string {
	return f.Name
}

func (f *IntFlag) AddToCmdFlag(cmds ...*cobra.Command) *IntFlag {
	for _, cmd := range cmds {
		cmd.Flags().IntVar(&f.Value, f.Name, f.Default, f.Usage)
		for _, alias := range f.Aliases {
			cmd.Flags().IntVar(&f.Value, alias, f.Default, fmt.Sprintf("\t[alias for '%s']", f.Name))
			_ = cmd.Flags().MarkHidden(alias)
		}
	}
	return f
}

func (f *IntFlag) AddToCmdPersistentFlags(cmds ...*cobra.Command) *IntFlag {
	for _, cmd := range cmds {
		cmd.PersistentFlags().IntVar(&f.Value, f.Name, f.Default, f.Usage)
		for _, alias := range f.Aliases {
			cmd.PersistentFlags().IntVar(&f.Value, alias, f.Default, fmt.Sprintf("\t[alias for '%s']", f.Name))
			_ = cmd.PersistentFlags().MarkHidden(alias)
		}
	}
	return f
}
