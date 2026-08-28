package cmd

// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) Lewis Cook <hi@lcook.net>

import (
	"os"

	"github.com/spf13/cobra"

	log "github.com/lcook/yorha/internal/logger"
	"github.com/lcook/yorha/internal/ostree"
)

var (
	force bool

	updateCmd = &cobra.Command{
		Use:   "update",
		Short: "Update the system using the latest container image",
		PreRun: func(cmd *cobra.Command, args []string) {
			if os.Getuid() != 0 {
				log.Error("Administrative privileges required")
			}

			if !ostree.Environment() {
				log.Error(
					"This command must be executed within an active OSTree environment",
				)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			ostreeConfig.SysSetup = "/var/tmp"
			ostreeConfig.SysTree = "/var/tmp/rootfs"
			ostreeConfig.ForceUpdate = force

			otree := ostree.New(ostreeConfig)
			if image == "" {
				deployments, err := otree.GetDeployments()
				if err != nil {
					log.Error(err.Error())
				}

				for _, deployment := range deployments {
					if deployment.Booted {
						release := deployment.OSRelease()

						if val, ok := release["IMAGE"]; ok {
							log.Infof(
								"Detected booted OSTree environment with active image: %s",
								val,
							)
							image = val
						}

						break
					}
				}
			}

			otree.Image = image

			otree.CreateRootFilesystem()
			otree.CreateLayout()
			otree.DeployImage()

			log.Info(
				"Update completed successfully. Reboot to activate the new deployment",
			)
		},
	}
)

func init() {
	updateCmd.Flags().
		StringVarP(&image, "Container image", "i", "", "Name of container image")
	updateCmd.Flags().
		BoolVarP(&force, "Force update", "f", false, "Force update")

	rootCmd.AddCommand(updateCmd)
}
