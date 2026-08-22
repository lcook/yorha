package cmd

// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) Lewis Cook <hi@lcook.net>

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	log "github.com/lcook/yorha/internal/logger"
	"github.com/lcook/yorha/internal/ostree"
)

var switchCmd = &cobra.Command{
	Use:   "switch",
	Short: "Switch the active OSTree deployment",
	PreRun: func(cmd *cobra.Command, args []string) {
		if os.Getuid() != 0 {
			log.Error("root permission required to run this operation")
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		otree := ostree.New(ostreeConfig)

		deployments, err := otree.GetDeployments()
		if err != nil {
			log.Error(err.Error())
		}

		if len(deployments) == 1 {
			log.Error("No other OSTree deployments to choose from")
		}

		otree.PrintDeployments()
		fmt.Println()

		input := log.Inputf(
			"Select deployment index [0-%d]: ",
			len(deployments)-1,
		)

		index, err := strconv.Atoi(input)
		if err != nil {
			log.Errorf(
				"'%s' is an invalid index [0-%d]",
				input,
				len(deployments)-1,
			)
		}

		if index > len(deployments)-1 || index < 0 {
			log.Errorf(
				"'%s' is an invalid index [0-%d]",
				input,
				len(deployments)-1,
			)
		}

		err = otree.SwitchDeployment(index)
		if err != nil {
			log.Errorf("Unable to switch deployment: %s", err.Error())
		}

		log.Infof("Deployment set to: %d", index)
	},
}

func init() {
	rootCmd.AddCommand(switchCmd)
}
