package cmd

// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) Lewis Cook <hi@lcook.net>

import (
	"github.com/spf13/cobra"

	"github.com/lcook/yorha/internal/ostree"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List current OSTree deployments",
	Run: func(cmd *cobra.Command, args []string) {
		ostree.PrintDeployments(opt)
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
