package disk

// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) Lewis Cook <hi@lcook.net>

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type state int

const (
	StateLive state = iota
	StateRunning
)

type DiskEntry struct {
	Name, Model, Path, State string
	Size                     float64
}

func (d DiskEntry) String() string {
	return fmt.Sprintf("%s\t%s\t%s\t%.2fGiB", d.Path, d.Model, d.State, d.Size)
}

func ReadSysBlockAttr(device, attr string) string {
	data, err := os.ReadFile(filepath.Join("/sys/block", device, attr))
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(data))
}

func GetDisks() ([]DiskEntry, error) {
	entries, err := os.ReadDir("/sys/block")
	if err != nil {
		return []DiskEntry{}, err
	}

	var disks []DiskEntry

	for _, dirent := range entries {
		name := dirent.Name()

		if strings.HasPrefix(name, "loop") ||
			strings.HasPrefix(name, "dm-") ||
			strings.HasPrefix(name, "ram") ||
			strings.HasPrefix(name, "sr") ||
			strings.HasPrefix(name, "zram") {
			continue
		}

		size, _ := strconv.ParseUint(ReadSysBlockAttr(name, "size"), 0, 64)
		sector, _ := strconv.ParseUint(
			ReadSysBlockAttr(name, "queue/hw_sector_size"),
			0,
			64,
		)

		disks = append(disks, DiskEntry{
			Name:  name,
			Model: ReadSysBlockAttr(name, "device/model"),
			State: ReadSysBlockAttr(name, "device/state"),
			Path:  filepath.Join("/dev", name),
			Size:  float64(sector*size) / (1024 * 1024 * 1024),
		})
	}

	return disks, nil
}
