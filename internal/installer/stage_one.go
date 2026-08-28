package installer

// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) Lewis Cook <hi@lcook.net>

import (
	"fmt"
	"strings"

	log "github.com/lcook/yorha/internal/logger"
)

func (i *Installer) WipeDisk() {
	err := log.Runf(
		[]string{"wipefs", "-a", i.Target.Path},
		"Preparing storage device (%s)",
		i.Target.Path,
	)
	if err != nil {
		log.Errorf("Failed to wipe disk: %s", err.Error())
	}
}

func (i *Installer) CreateLayout() {
	var command strings.Builder
	fmt.Fprintf(
		&command,
		"parted -a optimal -s %s -- mklabel gpt ",
		i.Target.Path,
	)

	command.WriteString("mkpart SYS_BOOT fat32 0% 500MiB ")
	command.WriteString("set 1 esp on ")

	fmt.Fprintf(&command, "mkpart SYS_ROOT xfs 500MiB %dGiB ", i.RootSize)
	fmt.Fprintf(&command, "mkpart SYS_VAR xfs %dGiB 100%%", i.RootSize)

	err := log.Runf(
		strings.Fields(command.String()),
		"Creating partition layout (%s): boot (500MiB) | root (%dGiB) | var (remaining space)",
		i.Target.Path,
		i.RootSize,
	)
	if err != nil {
		log.Errorf("Failed to create partition layout: %s", err.Error())
	}
}

func (i *Installer) CreateFormat() {
	err := log.Runf(
		[]string{
			"mkfs.vfat",
			"-n",
			"SYS_BOOT",
			"-F",
			"32",
			i.Manager.PartLabels["SYS_BOOT"],
		},
		"Formatting boot partition (%s)",
		i.Manager.PartLabels["SYS_BOOT"],
	)
	if err != nil {
		log.Errorf("Failed to format boot partition: %s", err.Error())
	}

	err = log.Runf(
		[]string{
			"mkfs.xfs",
			"-L",
			"SYS_ROOT",
			"-f",
			i.Manager.PartLabels["SYS_ROOT"],
			"-n",
			"ftype=1",
		},
		"Formatting root partition (%s)",
		i.Manager.PartLabels["SYS_ROOT"],
	)
	if err != nil {
		log.Errorf("Failed to format root partition: %s", err.Error())
	}

	err = log.Runf(
		[]string{
			"mkfs.xfs",
			"-L",
			"SYS_VAR",
			"-f",
			i.Manager.PartLabels["SYS_VAR"],
			"-n",
			"ftype=1",
		},
		"Formatting var partition (%s)",
		i.Manager.PartLabels["SYS_VAR"],
	)
	if err != nil {
		log.Errorf("Failed to format var partition: %s", err.Error())
	}
}
