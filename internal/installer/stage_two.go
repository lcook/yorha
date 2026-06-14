package installer

// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) Lewis Cook <hi@lcook.net>

import (
	"fmt"
	"os"
	"regexp"

	"golang.org/x/sys/unix"
)

func (i *Installer) CreateMounts() {
	if _, err := os.Stat(i.SysRoot); err != nil {
		if err := os.MkdirAll(i.SysRoot, 0o755); err != nil {
			i.Log.Errorf("Failed to create sysroot mount directory: %s", err.Error())
		}

		i.Log.Info(
			"Creating missing sysroot mount directory at " + i.SysRoot,
		)
	}

	i.Log.Infof(
		"Mounting root partition %s at %s",
		i.PartLabels["SYS_ROOT"],
		i.SysRoot,
	)

	if err := unix.Mount(
		i.PartLabels["SYS_ROOT"],
		i.SysRoot,
		"xfs",
		0,
		"",
	); err != nil {
		i.Log.Errorf("Failed to mount root partition at sysroot: %s", err.Error())
	}

	efiDir := fmt.Sprintf("%s/boot/efi", i.SysRoot)

	if _, err := os.Stat(efiDir); err != nil {
		if err := os.MkdirAll(efiDir, 0o755); err != nil {
			i.Log.Errorf("Failed to create EFI system partition mount point: %s", err.Error())
		}

		i.Log.Infof("Creating missing EFI system partition mount point at %s", efiDir)
	}

	i.Log.Infof(
		"Mounting boot partition %s at %s",
		i.PartLabels["SYS_BOOT"],
		efiDir,
	)

	if err := unix.Mount(
		i.PartLabels["SYS_BOOT"],
		efiDir,
		"vfat",
		uintptr(0),
		"",
	); err != nil {
		i.Log.Errorf("Unable to mount boot partition: %s", err.Error())
	}
}

func (i *Installer) CreateRepository() {
	i.Log.Run(
		"Initializing OSTree filesystem layout",
		[]string{
			"ostree",
			"admin",
			"init-fs",
			"--sysroot=/mnt",
			"--modern",
			"/mnt",
		},
	)

	i.Log.Run("Initializing OSTree stateroot", []string{
		"ostree",
		"admin",
		"stateroot-init",
		"--sysroot=/mnt",
		"yorha",
	})

	i.Log.Run("Initializing bare OSTree repository", []string{
		"ostree",
		"init",
		"--repo=/mnt/ostree/repo",
		"--mode=bare",
	})

	i.Log.Run(
		"Enabling relative boot paths for BLS entries",
		[]string{
			"ostree",
			"config",
			"--repo=/mnt/ostree/repo",
			"set",
			"sysroot.bootprefix",
			"true",
		},
	)
}

func (i *Installer) PatchStorage() {
	var (
		storage      = "/etc/containers/storage.conf"
		storageRegex = regexp.MustCompile(`(?m)^(graphroot\s*=\s*).*$`)

		containers      = "/etc/containers/containers.conf"
		containersRegex = regexp.MustCompile(
			`(?m)^# image_copy_tmp_dir\s*=\s*.*`,
		)
	)

	content, err := os.ReadFile(storage)
	if err != nil {
		i.Log.Errorf(
			"Failed to read container storage configuration %s: %s",
			storage,
			err.Error(),
		)
	}

	newContent := storageRegex.ReplaceAllString(
		string(content),
		fmt.Sprintf(`$1"%s/container-tmp"`, i.SysSetup),
	)

	err = os.WriteFile(storage, []byte(newContent), 0o644)
	if err != nil {
		i.Log.Errorf("Failed to write storage configuration %s: %s",
			storage,
			err.Error(),
		)
	}

	i.Log.Infof(
		"Configured image storage root to %s/container-storage",
		i.SysSetup,
	)

	content, err = os.ReadFile(containers)
	if err != nil {
		i.Log.Errorf(
			"Failed to read containers configuration %s: %s",
			containers,
			err.Error(),
		)
	}

	newContent = containersRegex.ReplaceAllString(
		string(content),
		fmt.Sprintf(
			`image_copy_tmp_dir = "%s/container-tmp"`,
			i.SysSetup,
		),
	)

	err = os.WriteFile(containers, []byte(newContent), 0o644)
	if err != nil {
		i.Log.Errorf(
			"Failed to write containers configuration %s: %s",
			containers,
			err.Error(),
		)
	}

	i.Log.Infof(
		"Configured temporary image staging directory to %s/container-tmp",
		i.SysSetup,
	)
}
