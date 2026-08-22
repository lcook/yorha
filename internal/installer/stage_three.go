package installer

// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) Lewis Cook <hi@lcook.net>

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"

	log "github.com/lcook/yorha/internal/logger"
	"github.com/lcook/yorha/internal/ostree"
)

func (i *Installer) InstallBootloader() {
	log.Run(
		"Installing GRUB bootloader for OSTree deployment",
		[]string{
			"grub-install",
			"--target=x86_64-efi",
			fmt.Sprintf(
				"--efi-directory=%s/boot/efi",
				i.Manager.SysRoot,
			),
			"--removable",
			fmt.Sprintf(
				"--boot-directory=%s/boot/efi/EFI",
				i.Manager.SysRoot,
			),
			"--bootloader-id=" + ostree.DefaultStateroot,
			i.Manager.PartLabels["SYS_BOOT"],
		},
	)

	deployments, _ := i.Manager.GetDeployments()

	syspath := fmt.Sprintf(
		"%s/ostree/deploy/%s/deploy/%s.0",
		i.Manager.SysRoot,
		ostree.DefaultStateroot,
		deployments[0].Checksum,
	)

	matches, _ := filepath.Glob(
		filepath.Join(syspath, "boot", "*"),
	)
	for _, m := range matches {
		os.RemoveAll(m)
	}

	if err := unix.Mount(
		i.Manager.SysRoot+"/boot",
		syspath+"/boot",
		"xfs",
		uintptr(unix.MS_BIND|unix.MS_REC),
		"",
	); err != nil {
		log.Errorf(
			"Failed to bind-mount boot directory into OSTree deployment path for GRUB installation: %s",
			err.Error(),
		)
	}

	os.MkdirAll(syspath+"/sysroot/ostree", 0o755)

	if err := unix.Mount(
		i.Manager.SysRoot+"/ostree",
		syspath+"/sysroot/ostree",
		"xfs",
		uintptr(unix.MS_BIND|unix.MS_REC),
		"",
	); err != nil {
		log.Errorf(
			"Failed to bind-mount OSTree directory into GRUB installation chroot: %s",
			err.Error(),
		)
	}

	for _, dev := range []string{"/dev", "/proc", "/sys"} {
		if err := unix.Mount(
			dev,
			syspath+dev,
			"",
			uintptr(unix.MS_BIND),
			"",
		); err != nil {
			log.Errorf(
				"Failed to bind-mount %s into GRUB installation chroot: %s",
				dev,
				err.Error(),
			)
		}
	}

	log.Run(
		"Generating GRUB configuration inside for new OSTree deployment",
		[]string{
			"chroot",
			syspath,
			"/bin/bash",
			"-c",
			"grub-mkconfig -o /boot/efi/EFI/grub/grub.cfg",
		},
	)

	os.RemoveAll(i.Manager.SysSetup)
}
