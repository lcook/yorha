package installer

// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) Lewis Cook <hi@lcook.net>

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	log "github.com/lcook/yorha/internal/logger"
	"golang.org/x/sys/unix"
)

func (i *Installer) CreateMounts() {
	if _, err := os.Stat(i.Manager.SysRoot); err != nil {
		if err := os.MkdirAll(i.Manager.SysRoot, 0o755); err != nil {
			log.Errorf(
				"Failed to create sysroot mount directory: %s",
				err.Error(),
			)
		}

		log.Infof(
			"Creating missing sysroot mount directory at %s",
			i.Manager.SysRoot,
		)
	}

	log.Infof(
		"Mounting root partition %s at %s",
		i.Partitions.Root,
		i.Manager.SysRoot,
	)

	if err := unix.Mount(
		i.Partitions.Root,
		i.Manager.SysRoot,
		"xfs",
		0,
		"",
	); err != nil {
		log.Errorf(
			"Failed to mount root partition at sysroot: %s",
			err.Error(),
		)
	}

	efiDir := fmt.Sprintf("%s/boot/efi", i.Manager.SysRoot)

	if _, err := os.Stat(efiDir); err != nil {
		if err := os.MkdirAll(efiDir, 0o755); err != nil {
			log.Errorf(
				"Failed to create EFI system partition mount point: %s",
				err.Error(),
			)
		}

		log.Infof(
			"Creating missing EFI system partition mount point at %s",
			efiDir,
		)
	}

	log.Infof(
		"Mounting boot partition %s at %s",
		i.Partitions.Boot,
		efiDir,
	)

	if err := unix.Mount(
		i.Partitions.Boot,
		efiDir,
		"vfat",
		uintptr(0),
		"",
	); err != nil {
		log.Errorf("Unable to mount boot partition: %s", err.Error())
	}
}

func (i *Installer) CreateRepository() {
	log.Run(
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

	log.Run("Initializing OSTree stateroot", []string{
		"ostree",
		"admin",
		"stateroot-init",
		"--sysroot=/mnt",
		"yorha",
	})

	log.Run("Initializing bare OSTree repository", []string{
		"ostree",
		"init",
		"--repo=/mnt/ostree/repo",
		"--mode=bare",
	})

	log.Run(
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
		storage      = "/usr/share/containers/storage.conf"
		storageRegex = regexp.MustCompile(`(?m)^# (graphroot\s*=\s*).*$`)

		containers      = "/usr/share/containers/containers.conf"
		containersRegex = regexp.MustCompile(
			`(?m)^# image_copy_tmp_dir\s*=\s*.*`,
		)
	)

	content, err := os.ReadFile(storage)
	if err != nil {
		log.Errorf(
			"Failed to read container storage configuration %s: %s",
			storage,
			err.Error(),
		)
	}

	newContent := storageRegex.ReplaceAllString(
		string(content),
		fmt.Sprintf(`$1"%s/container-tmp"`, i.Manager.SysSetup),
	)

	err = os.WriteFile(storage, []byte(newContent), 0o644)
	if err != nil {
		log.Errorf(
			"Failed to write storage configuration %s: %s",
			storage,
			err.Error(),
		)
	}

	log.Infof(
		"Configured image storage root to %s/container-storage",
		i.Manager.SysSetup,
	)

	os.MkdirAll(filepath.Join(i.Manager.SysSetup, "container-storage"), 0o755)

	content, err = os.ReadFile(containers)
	if err != nil {
		log.Errorf(
			"Failed to read containers configuration %s: %s",
			containers,
			err.Error(),
		)
	}

	newContent = containersRegex.ReplaceAllString(
		string(content),
		fmt.Sprintf(
			`image_copy_tmp_dir = "%s/container-tmp"`,
			i.Manager.SysSetup,
		),
	)

	err = os.WriteFile(containers, []byte(newContent), 0o644)
	if err != nil {
		log.Errorf(
			"Failed to write containers configuration %s: %s",
			containers,
			err.Error(),
		)
	}

	log.Infof(
		"Configured temporary image staging directory to %s/container-tmp",
		i.Manager.SysSetup,
	)

	os.MkdirAll(filepath.Join(i.Manager.SysSetup, "container-tmp"), 0o755)
}
