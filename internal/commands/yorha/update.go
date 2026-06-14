package cmd

// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) Lewis Cook <hi@lcook.net>

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/lcook/yorha/internal/ostree"
)

var (
	force bool

	updateCmd = &cobra.Command{
		Use:   "update",
		Short: "Update the system using the latest container image",
		PreRun: func(cmd *cobra.Command, args []string) {
			if os.Getuid() != 0 {
				opt.Log.Error("root permission required to run this operation")
			}

			if !ostree.Environment() {
				opt.Log.Error(
					"Update aborted, this operation must be run inside an active OSTree environment",
				)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			opt.SysSetup = "/var/tmp"
			opt.SysTree = "/var/tmp/rootfs"
			opt.ForceUpdate = force

			if image == "" {
				deployments, err := ostree.GetDeployments(opt)
				if err != nil {
					opt.Log.Error(err.Error())
				}

				for _, deployment := range deployments {
					if deployment.Booted {
						release := deployment.OSRelease()

						if val, ok := release["IMAGE"]; ok {
							opt.Log.Infof(
								"Detected booted OSTree environment with image %s",
								val,
							)
							image = val
						}

						break
					}
				}
			}

			opt.Image = image

			ostree.CreateRootFilesystem(opt)
			ostree.CreateLayout(opt)
			ostree.DeployImage(opt)

			opt.Log.Info("Update complete. Reboot to use the new deployment")
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
