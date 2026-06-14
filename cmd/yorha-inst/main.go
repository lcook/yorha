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
	"github.com/lcook/yorha/internal/logger"
	"github.com/lcook/yorha/internal/version"
)

var deps = []string{"podman", "ostree"}

func main() {
	log := logger.New()

	color.Yellow(`yorha installer (ver:%s)`, version.Build)
	fmt.Println()

	for _, dep := range deps {
		_, err := exec.LookPath(dep)
		if err != nil {
			log.Errorf(
				"Required dependency '%s' not found on system, install the missing package and try again",
				dep,
			)
		}
	}

	disks, err := disk.GetDisks()
	if err != nil {
		log.Errorf("Failed to probe disks: %s", err.Error())
	}

	if len(disks) == 0 {
		log.Error("No usable disks detected on system")
	}

	idx := slices.IndexFunc(
		disks,
		func(d disk.DiskEntry) bool { return d.State == "live" },
	)

	if idx < 0 {
		idx = 0
	}

	var (
		inst                        *installer.Installer
		partboot, partroot, partvar string
		size                        string
	)

	log.Info("Detected available disks:")

	for _, disk := range disks {
		fmt.Printf("  %s\n", disk)
	}

	for {
		selected := log.Inputf(
			"Enter the disk device to format ('?' for details) [%s]: ",
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
			log.Warnf("'%s' is not a valid disk", dev)

			continue
		}

		selected = log.Inputf("Confirm selected disk '%s' [y/N] ", dev)

		answer := strings.ToLower(selected)
		if answer == "" || (answer != "y" && answer != "yes") {
			log.Warn("Disk selection cancelled")
			continue
		}

		log.Warnf("This will permanently erase all data on '%s'", dev)

		selected = log.Inputf(
			"Proceed with formatting '%s'? [y/N] ",
			dev,
		)

		answer = strings.ToLower(selected)
		if answer == "" || (answer != "y" && answer != "yes") {
			log.Warn("Operation aborted by user")

			continue
		}

		if strings.Contains(dev, "nvme") ||
			strings.Contains(dev, "mmcblk") {
			partboot = dev + "p1"
			partroot = dev + "p2"
			partvar = dev + "p3"
		} else {
			partboot = dev + "1"
			partroot = dev + "2"
			partvar = dev + "3"
		}

		i := slices.IndexFunc(
			disks,
			func(d disk.DiskEntry) bool { return d.Path == dev },
		)

		selected = log.Input(
			"Enter target root partition size in GiB [25]: ")

		size = strings.TrimPrefix(
			strings.TrimSpace(strings.ToLower(selected)),
			"gib",
		)

		if size == "" {
			size = "25"
		}

		rootsize, err := strconv.Atoi(size)
		if err != nil {
			log.Errorf("Invalid root partition size: %s", err.Error())
		}

		inst = installer.New(disks[i], partboot, partroot, partvar, rootsize)

		break
	}

	go func() {
		sc := make(chan os.Signal, 1)
		signal.Notify(sc, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

		switch <-sc {
		case os.Interrupt, syscall.SIGINT, syscall.SIGTERM:
			fmt.Println()
			log.Warn("Operation aborted: exiting installer")

			if err := unix.Unmount(inst.SysRoot, unix.MNT_DETACH); err == nil {
				log.Infof("Unmounted sysroot directory: %s", inst.SysRoot)
			}

			os.Exit(0)
		}
	}()

	inst.Run()

	log.Info(
		"Installation complete. Reboot into your new system with `systemctl reboot`",
	)
}
