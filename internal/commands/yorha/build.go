//go:build !thin

package cmd

// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) Lewis Cook <hi@lcook.net>

import (
	"github.com/spf13/cobra"

	"github.com/lcook/yorha/internal/config"
	"github.com/lcook/yorha/internal/podman"
)

var (
	rootfull bool

	buildCmd = &cobra.Command{
		Use:   "build",
		Short: "Build container from configuration file",
		Run: func(cmd *cobra.Command, args []string) {
			socket := podman.RootlessContext
			if rootfull {
				socket = podman.RootfullContext
			}

			podman, err := podman.NewClient(socket)
			if err != nil {
				opt.Log.Error(err.Error())
			}

			config, err := config.FromFile[config.Container](configfile)
			if err != nil {
				opt.Log.Error(err.Error())
			}

			if !podman.HasLocalImage(config.Depends) {
				opt.Log.Errorf(
					"Required container image dependency '%s' is missing from local storage when building '%s'",
					config.Depends,
					config.Image,
				)
			}

			err = podman.BuildContainer(config)
			if err != nil {
				opt.Log.Error(err.Error())
			}
		},
	}
)

func init() {
	buildCmd.Flags().
		StringVarP(&configfile, "config", "c", "config-base.yaml", "Path to configuration file")
	buildCmd.Flags().
		BoolVarP(&rootfull, "rootfull", "r", false, "Build in rootfull context")

	rootCmd.AddCommand(buildCmd)
}
