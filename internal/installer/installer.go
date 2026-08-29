package installer

// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) Lewis Cook <hi@lcook.net>

import (
	"golang.org/x/sys/unix"

	"github.com/lcook/yorha/internal/disk"
	log "github.com/lcook/yorha/internal/logger"
	"github.com/lcook/yorha/internal/ostree"
)

const (
	DefaultImage = "ghcr.io/lcook/yorha/archlinux-mainline"
)

type Partitions struct {
	Boot     string
	Root     string
	Var      string
	RootSize int
}

type Installer struct {
	Target     disk.DiskEntry
	Manager    *ostree.Manager
	Partitions Partitions
}

func New(
	disk disk.DiskEntry,
	partitions Partitions,
) *Installer {
	return &Installer{
		Target: disk,
		Manager: ostree.New(
			ostree.Config{
				SysRoot:     "/mnt",
				SysSetup:    "/mnt/setup",
				SysTree:     "/mnt/setup/root",
				Image:       DefaultImage,
				Interactive: true,
			},
		),
		Partitions: partitions,
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

	log.Info("Phase one complete: Storage prepared")
}

func (i *Installer) StageTwo() {
	i.CreateMounts()
	i.CreateRepository()
	i.PatchStorage()

	i.Manager.CreateRootFilesystem()
	i.Manager.CreateLayout()

	log.Info("Phase two complete: OSTree repository setup and image staged")
}

func (i *Installer) StageThree() {
	i.Manager.DeployImage()

	i.InstallBootloader()

	if err := unix.Unmount(i.Manager.SysRoot, unix.MNT_DETACH); err != nil {
		log.Errorf("Failed to unmount sysroot: %s", err.Error())
	}

	log.Info(
		"Phase three complete: System deployed and bootloader configured",
	)
}
