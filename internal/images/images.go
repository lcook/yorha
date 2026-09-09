package images

// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) Lewis Cook <hi@lcook.net>

var (
	Images = []struct {
		Name        string
		Description string
	}{
		{
			"ghcr.io/lcook/yorha/archlinux-base",
			"YoRHa base image with system tooling",
		},
		{
			"ghcr.io/lcook/yorha/archlinux-mainline",
			"YoRHa desktop image with Hyprland",
		},
		{
			"ghcr.io/lcook/yorha/archlinux-nvidia",
			"YoRHa desktop image with NVIDIA graphics drivers",
		},
		{
			"ghcr.io/lcook/yorha/archlinux-intel",
			"YoRHa desktop image with Intel graphics drivers and media acceleration",
		},
		{
			"Custom",
			"Specify location to your own image",
		},
	}
	DefaultImage = Images[0].Name
)
