package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spidernet-io/rocktemplate/pkg/ebpf"
	"os"
)

var CmdPrintMapNodeIp = &cobra.Command{
	Use:   "nodeIp",
	Short: "print the ebpf map of nodeIp ",
	Args:  cobra.RangeArgs(0, 0),
	Run: func(cmd *cobra.Command, args []string) {
		bpf := ebpf.NewEbpfProgramMananger(nil)
		if err := bpf.LoadAllEbpfMap(""); err != nil {
			fmt.Printf("failed to load ebpf Map: %v\n", err)
			os.Exit(2)
		}
		defer bpf.UnloadAllEbpfMap()

		fmt.Printf("\n")
		fmt.Printf("print the ebpf map of nodeIp:\n")
		bpf.PrintMapNodeIp()
		fmt.Printf("\n")
	},
}

func init() {
	CmdPrintMap.AddCommand(CmdPrintMapNodeIp)
}