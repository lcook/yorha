//go:build !thin

package cmd

// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) Lewis Cook <hi@lcook.net>

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/lcook/yorha/internal/config"
	log "github.com/lcook/yorha/internal/logger"
	"github.com/lcook/yorha/internal/podman"
)

var (
	rootfull bool

	buildCmd = &cobra.Command{
		Use:   "build",
		Short: "Build container from configuration file",
		PreRun: func(cmd *cobra.Command, args []string) {
			if os.Getuid() == 0 {
				rootfull = true
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			socket := podman.RootlessContext
			if rootfull {
				socket = podman.RootfullContext
			}

			podman, err := podman.NewClient(socket)
			if err != nil {
				log.Error(err.Error())
			}

			config, err := config.FromFile[config.Container](configfile)
			if err != nil {
				log.Error(err.Error())
			}

			if !podman.HasLocalImage(config.Depends) {
				log.Errorf(
					"Required dependency '%s' not available for building '%s'",
					config.Depends,
					config.Image,
				)
			}

			err = podman.BuildContainer(config)
			if err != nil {
				log.Error(err.Error())
			}
		},
	}
)

func init() {
	buildCmd.Flags().
		StringVarP(&configfile, "config", "c", "config-base.yaml", "Path to configuration file")
	buildCmd.MarkFlagRequired("config")

	rootCmd.AddCommand(buildCmd)
}
