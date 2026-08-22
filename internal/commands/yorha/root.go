package cmd

// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) Lewis Cook <hi@lcook.net>

import (
	"os"

	"github.com/lcook/yorha/internal/ostree"
	"github.com/spf13/cobra"
)

var (
	image        string
	output       string
	ostreeConfig = ostree.Config{SysRoot: "/"}
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
