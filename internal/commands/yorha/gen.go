//go:build !thin

package cmd

// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) Lewis Cook <hi@lcook.net>

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/lcook/yorha/internal/config"
	"github.com/lcook/yorha/internal/util"
)

var (
	configfile string
	configdir  string
	list       bool

	genCmd = &cobra.Command{
		Use:   "gen",
		Short: "Generate a Podman Containerfile from YAML configuration",
		Run: func(cmd *cobra.Command, args []string) {
			if list && configdir != "" {
				_, err := os.ReadDir(configdir)
				if err != nil {
					opt.Log.Error(err.Error())
				}

				configs := make(map[string]config.Container)

				filepath.Walk(
					configdir,
					func(path string, info fs.FileInfo, err error) error {
						if err != nil {
							opt.Log.Error(err.Error())
						}

						if info.IsDir() {
							return nil
						}

						c, err := config.FromFile[config.Container](path)
						if err != nil {
							opt.Log.Error(err.Error())
						}

						configs[info.Name()] = c

						return nil
					},
				)

				if len(configs) == 0 {
					opt.Log.Errorf(
						"No valid configurations found in %s",
						configdir,
					)
				}

				writer := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
				fmt.Fprintln(writer, "Configuration\tImage\tDepends\tComment")

				for file, container := range configs {
					fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
						file,
						container.Image,
						container.Depends,
						container.Comment,
					)
				}

				writer.Flush()
			}

			if configdir != "" {
				configfile = filepath.Join(configdir, configfile)
			}

			config, err := config.FromFile[config.Container](configfile)
			if err != nil {
				opt.Log.Error(err.Error())
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

			err = util.GenerateContainerfile(handle, config)
			if err != nil {
				opt.Log.Error(err.Error())
			}
		},
	}
)

func init() {
	genCmd.Flags().
		StringVarP(&configfile, "config", "c", "", "Path to configuration file")
	genCmd.Flags().
		StringVarP(&output, "output", "o", "", "Output Containerfile")
	genCmd.Flags().
		BoolVarP(&list, "list", "l", false, "List found configuration files")
	genCmd.Flags().
		StringVarP(&configdir, "dir", "d", "", "Directory containing configuration files")

	rootCmd.AddCommand(genCmd)
}
