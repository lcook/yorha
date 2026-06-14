package cmd

// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) Lewis Cook <hi@lcook.net>

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/lcook/yorha/internal/logger"
	"github.com/lcook/yorha/internal/ostree"
)

var (
	image  string
	output string
	opt    = &ostree.Options{
		SysRoot: "/",
		Log:     logger.New(),
	}
)

var rootCmd = &cobra.Command{
	CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true},
	SilenceUsage:      true,
	Use:               "yorha",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
