//go:build !thin

package cmd

// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) Lewis Cook <hi@lcook.net>

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/lcook/yorha/internal/podman"
	"github.com/lcook/yorha/internal/util"
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export a local container image to a tar archive",
	Run: func(cmd *cobra.Command, args []string) {
		podman, err := podman.NewClient(podman.RootlessContext)
		if err != nil {
			opt.Log.Error(err.Error())
		}

		if !podman.HasLocalImage(image) {
			opt.Log.Errorf(
				"Container image '%s' not found in local storage",
				image,
			)
		}

		handle, err := util.GetFileDescriptor(output)
		if err != nil {
			opt.Log.Error(err.Error())
		}

		defer func() {
			if handle != os.Stdout {
				handle.Close()
			}
		}()

		err = podman.ExportContainer(image, handle)
		if err != nil {
			opt.Log.Error(err.Error())
		}
	},
}

func init() {
	exportCmd.Flags().StringVarP(&image, "image", "i", "", "Image name")
	exportCmd.Flags().StringVarP(&output, "output", "o", "", "Output")

	exportCmd.MarkFlagRequired("image")
	exportCmd.MarkFlagRequired("output")

	rootCmd.AddCommand(exportCmd)
}
