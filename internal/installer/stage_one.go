package installer

// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) Lewis Cook <hi@lcook.net>

import (
	"fmt"
	"strings"
)

func (i *Installer) WipeDisk() {
	err := i.Log.Runf(
		[]string{"wipefs", "-a", i.Target.Path},
		"Wiping disk %s: all existing data on the device will be permanently erased",
		i.Target.Path,
	)
	if err != nil {
		i.Log.Errorf("Failed to wipe disk: %s", err.Error())
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

	err := i.Log.Runf(
		strings.Fields(command.String()),
		"Creating GPT partition layout on %s (SYS_BOOT: 500MiB, SYS_ROOT: %dGiB, SYS_VAR: remaining space)",
		i.Target.Path,
		i.RootSize,
	)
	if err != nil {
		i.Log.Errorf("Failed to create partition layout: %s", err.Error())
	}
}

func (i *Installer) CreateFormat() {
	err := i.Log.Runf(
		[]string{
			"mkfs.vfat",
			"-n",
			"SYS_BOOT",
			"-F",
			"32",
			i.PartLabels["SYS_BOOT"],
		},
		"Formatting EFI boot partition (%s) as FAT32",
		i.PartLabels["SYS_BOOT"],
	)
	if err != nil {
		i.Log.Errorf("Failed to format boot partition: %s", err.Error())
	}

	err = i.Log.Runf(
		[]string{
			"mkfs.xfs",
			"-L",
			"SYS_ROOT",
			"-f",
			i.PartLabels["SYS_ROOT"],
			"-n",
			"ftype=1",
		},
		"Formatting root partition (%s) as XFS",
		i.PartLabels["SYS_ROOT"],
	)
	if err != nil {
		i.Log.Errorf("Failed to format root partition: %s", err.Error())
	}

	err = i.Log.Runf(
		[]string{
			"mkfs.xfs",
			"-L",
			"SYS_VAR",
			"-f",
			i.PartLabels["SYS_VAR"],
			"-n",
			"ftype=1",
		},
		"Formatting /var partition (%s) as XFS",
		i.PartLabels["SYS_VAR"],
	)
	if err != nil {
		i.Log.Errorf("Failed to format var partition: %s", err.Error())
	}
}
