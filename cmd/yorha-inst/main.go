package main

// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) Lewis Cook <hi@lcook.net>

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"syscall"

	"github.com/fatih/color"
	"golang.org/x/sys/unix"

	"github.com/lcook/yorha/internal/disk"
	"github.com/lcook/yorha/internal/installer"
	log "github.com/lcook/yorha/internal/logger"
	"github.com/lcook/yorha/internal/version"
)

var deps = []string{"podman", "ostree"}

func main() {
	color.Yellow(`yorha installer (ver:%s)`, version.Build)
	fmt.Println()

	for _, dep := range deps {
		_, err := exec.LookPath(dep)
		if err != nil {
			log.Errorf(
				"Required dependency '%s' not found. Please install the package and try again",
				dep,
			)
		}
	}

	disks, err := disk.GetDisks()
	if err != nil {
		log.Errorf("Unable to enumerate storage devices: %s", err.Error())
	}

	if len(disks) == 0 {
		log.Error("No compatible storage devices detected")
	}

	idx := slices.IndexFunc(
		disks,
		func(d disk.DiskEntry) bool { return d.State == "live" },
	)

	if idx < 0 {
		idx = 0
	}

	var (
		inst *installer.Installer
		size string
	)

	log.Info("Available storage devices:")

	for _, disk := range disks {
		fmt.Printf("  %s\n", disk)
	}

	for {
		selected := log.Inputf(
			"Select target storage device ('?' for details) [%s]: ",
			disks[idx].Name,
		)

		if selected == "" {
			selected = disks[idx].Name
		}

		if selected == "?" {
			for _, disk := range disks {
				fmt.Printf("  %s\n", disk)
			}

			continue
		}

		dev := strings.ToLower(selected)
		if !strings.HasPrefix(selected, "/dev") {
			dev = filepath.Join("/dev", selected)
		}

		if !slices.ContainsFunc(
			disks,
			func(d disk.DiskEntry) bool { return d.Path == dev },
		) {
			log.Warnf("Device '%s' is not valid", dev)

			continue
		}

		selected = log.Inputf("Confirm selected disk '%s' [y/N] ", dev)

		answer := strings.ToLower(selected)
		if answer == "" || (answer != "y" && answer != "yes") {
			log.Warn("Device selection cancelled")
			continue
		}

		log.Warnf("Warning: All data on '%s' will be permanently erased", dev)

		selected = log.Inputf(
			"Continue with installation on '%s'? [y/N] ",
			dev,
		)

		answer = strings.ToLower(selected)
		if answer == "" || (answer != "y" && answer != "yes") {
			log.Warn("Installation cancelled")

			continue
		}

		var partitions installer.Partitions
		if strings.Contains(dev, "nvme") ||
			strings.Contains(dev, "mmcblk") {
			partitions.Boot = dev + "p1"
			partitions.Root = dev + "p2"
			partitions.Var = dev + "p3"
		} else {
			partitions.Boot = dev + "1"
			partitions.Root = dev + "2"
			partitions.Var = dev + "3"
		}

		i := slices.IndexFunc(
			disks,
			func(d disk.DiskEntry) bool { return d.Path == dev },
		)

		selected = log.Input(
			"Specify root partition size in GiB [30]: ",
		)

		size = strings.TrimPrefix(
			strings.TrimSpace(strings.ToLower(selected)),
			"gib",
		)

		if size == "" {
			size = "30"
		}

		rootsize, err := strconv.Atoi(size)
		if err != nil {
			log.Errorf("Invalid root partition size: %s", err.Error())
		}

		partitions.RootSize = rootsize

		inst = installer.New(disks[i], partitions)

		break
	}

	go func() {
		sc := make(chan os.Signal, 1)
		signal.Notify(sc, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

		switch <-sc {
		case os.Interrupt, syscall.SIGINT, syscall.SIGTERM:
			fmt.Println()
			log.Warn("Installation aborted: exiting installer")

			if err := unix.Unmount(
				inst.Manager.SysRoot,
				unix.MNT_DETACH,
			); err == nil {
				log.Infof(
					"Unmounted sysroot directory: %s",
					inst.Manager.SysRoot,
				)
			}

			os.Exit(0)
		}
	}()

	inst.Run()

	log.Info(
		"Installation completed successfully. Reboot with: systemctl reboot",
	)
}
