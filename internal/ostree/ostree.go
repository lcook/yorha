package ostree

// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) Lewis Cook <hi@lcook.net>

import (
	"os"
)

const (
	DefaultStateroot = "yorha"
	DefaultBranch    = "yorha/x86_64/standard"
)

type Config struct {
	SysRoot     string
	SysSetup    string
	SysTree     string
	Image       string
	PartLabels  map[string]string
	Interactive bool
	ForceUpdate bool
}

func Environment() bool {
	if _, err := os.Stat("/run/ostree-booted"); err != nil {
		return false
	}

	return true
}
