package config

// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) Lewis Cook <hi@lcook.net>

type Container struct {
	Image       string            `yaml:"image"`
	Comment     string            `yaml:"comment"`
	Environment map[string]string `yaml:"environment"`
	Depends     string            `yaml:"depends"`
	Packages    []string          `yaml:"packages"`
	Services    struct {
		Enable  []string `yaml:"enable"`
		Disable []string `yaml:"disable"`
		Mask    []string `yaml:"mask"`
	} `yaml:"services"`
	Files []struct {
		Source string `yaml:"source"`
		Dest   string `yaml:"dest"`
	} `yaml:"files"`
	Template string `yaml:"template"`
}
