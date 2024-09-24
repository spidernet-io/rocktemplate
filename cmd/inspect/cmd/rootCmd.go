// Copyright 2022 Authors of spidernet-io
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"os"
	"path/filepath"
)

var BinName = filepath.Base(os.Args[0])
var rootLogger *zap.Logger

var RootCmd = &cobra.Command{
	Use:   BinName,
	Short: "cli for debugging",
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		panic(err.Error())
	}
}
