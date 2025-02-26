package service

import (
	"github.com/spf13/cobra"
)

var ServiceCmd = &cobra.Command{
	Use:   "service",
	Short: "Service operations",
	Long:  `Service operations`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

func init() {
}
