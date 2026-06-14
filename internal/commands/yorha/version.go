package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/lcook/yorha/internal/version"
)

// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) Lewis Cook <hi@lcook.net>

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show toolkit version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version.Build)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
