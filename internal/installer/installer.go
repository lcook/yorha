package installer

// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) Lewis Cook <hi@lcook.net>

import (
	"golang.org/x/sys/unix"

	"github.com/lcook/yorha/internal/disk"
	"github.com/lcook/yorha/internal/logger"
	"github.com/lcook/yorha/internal/ostree"
)

const (
	DefaultImage = "ghcr.io/lcook/yorha/archlinux-mainline"
)

type Installer struct {
	Target disk.DiskEntry
	*ostree.Options
	RootSize int
}

func New(
	disk disk.DiskEntry,
	partboot, partroot, partvar string,
	rootsize int,
) *Installer {
	return &Installer{
		Target: disk,
		Options: &ostree.Options{
			SysRoot:  "/mnt",
			SysSetup: "/mnt/setup",
			SysTree:  "/mnt/setup/root",
			Image:    DefaultImage,
			PartLabels: map[string]string{
				"SYS_BOOT": partboot,
				"SYS_ROOT": partroot,
				"SYS_VAR":  partvar,
			},
			Interactive: true,
			Log:         logger.New(),
		},
		RootSize: rootsize,
	}
}

func (i *Installer) Run() {
	i.StageOne()
	i.StageTwo()
	i.StageThree()
}

func (i *Installer) StageOne() {
	i.WipeDisk()
	i.CreateLayout()
	i.CreateFormat()

	i.Log.Info("Stage one complete: disk partitioning and formatting")
}

func (i *Installer) StageTwo() {
	i.CreateMounts()
	i.CreateRepository()
	i.PatchStorage()

	ostree.CreateRootFilesystem(i.Options)
	ostree.CreateLayout(i.Options)

	i.Log.Info("Stage two complete: OSTree repository setup and image staging")
}

func (i *Installer) StageThree() {
	ostree.DeployImage(i.Options)

	i.InstallBootloader()

	if err := unix.Unmount(
		i.SysRoot,
		unix.MNT_DETACH,
	); err != nil {
		i.Log.Errorf("Failed to unmount sysroot: %s", err.Error())
	}

	i.Log.Info(
		"Stage three complete: OSTree deployment and bootloader installation",
	)
}
