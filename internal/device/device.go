package device

import (
	"fmt"
	"os"
	"slices"
	"strings"
)

// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) Lewis Cook <hi@lcook.net>

type Device int

const (
	DeviceUnknown Device = iota
	DeviceIntel
	DeviceNVIDIA

	VendorIntel  = "0x8086"
	VendorNVIDIA = "0x10de"
)

func ToStr(dev Device) string {
	switch dev {
	case DeviceIntel:
		return "Intel"
	case DeviceNVIDIA:
		return "NVIDIA"
	}

	return "Unknown"
}

func Detect() Device {
	dir, _ := os.ReadDir("/sys/class/drm/")

	var cards []string

	for _, dirent := range dir {
		name := dirent.Name()

		if strings.HasPrefix(name, "card") &&
			!strings.Contains(name, "-") {
			cards = append(cards, name)
		}
	}

	var devices []Device

	for _, card := range cards {
		vendor, _ := os.ReadFile(
			fmt.Sprintf("/sys/class/drm/%s/device/vendor", card),
		)

		switch strings.TrimSpace(string(vendor)) {
		case VendorIntel:
			devices = append(devices, DeviceIntel)
		case VendorNVIDIA:
			devices = append(devices, DeviceNVIDIA)
		}
	}

	preffered := DeviceUnknown

	if slices.Contains(devices, DeviceNVIDIA) {
		preffered = DeviceNVIDIA
	} else if slices.Contains(devices, DeviceIntel) {
		preffered = DeviceIntel
	}

	return preffered
}
