package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var CmdPrintMapAll = &cobra.Command{
	Use:   "all",
	Short: "print the all ebpf map ",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("\n")
		fmt.Printf("print the all ebpf map:\n")
	},
}

func init() {
	CmdPrintMap.AddCommand(CmdPrintMapAll)
}
