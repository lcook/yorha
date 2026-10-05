package images

import "github.com/lcook/yorha/internal/device"

// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) Lewis Cook <hi@lcook.net>

type Image struct {
	Name        string
	Description string
}

var Images = map[device.Device]Image{
	device.DeviceUnknown: {
		"ghcr.io/lcook/yorha/archlinux-mainline",
		"YoRHa desktop image with Hyprland",
	},
	device.DeviceIntel: {
		"ghcr.io/lcook/yorha/archlinux-intel",
		"YoRHa desktop image with Intel graphics drivers and media acceleration",
	},
	device.DeviceNVIDIA: {
		"ghcr.io/lcook/yorha/archlinux-nvidia",
		"YoRHa desktop image with NVIDIA graphics drivers",
	},
}

var DefaultImage = Images[device.DeviceUnknown].Name
