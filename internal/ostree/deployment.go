package ostree

// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) Lewis Cook <hi@lcook.net>

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

type Deployment struct {
	Checksum     string `json:"checksum"`
	Stateroot    string `json:"stateroot"`
	RefSpec      string `json:"refspec,omitempty"`
	Serial       int    `json:"serial"`
	Index        int    `json:"index"`
	Booted       bool   `json:"booted"`
	Pending      bool   `json:"pending"`
	Rollback     bool   `json:"rollback"`
	Finalization bool   `json:"finalization-locked"`
	SoftReboot   bool   `json:"soft-reboot-target"`
	Staged       bool   `json:"staged"`
	Pinned       bool   `json:"pinned"`
	Unlocked     string `json:"unlocked"`
	Version      string `json:"version,omitempty"`
}

func (d Deployment) Path() string {
	return fmt.Sprintf(
		"/ostree/deploy/%s/deploy/%s.0",
		d.Stateroot,
		d.Checksum,
	)
}

func (d Deployment) Created() string {
	var (
		created string
		release = d.OSRelease()
	)

	if val, ok := release["VERSION_ID"]; ok {
		version, err := time.ParseInLocation(
			"20060102.1504",
			strings.Split(val, "-")[0],
			time.Local,
		)
		if err != nil {
			return created
		}

		var (
			duration = time.Since(version)

			days    = int(duration.Hours() / 24)
			hours   = int(duration.Hours()) % 24
			minutes = int(duration.Minutes()) % 60
		)

		switch {
		case days > 0:
			if days == 1 {
				created = "1 day ago"
			} else {
				created = fmt.Sprintf("%d days ago", days)
			}

		case hours > 0:
			if hours == 1 {
				created = "1 hour ago"
			} else {
				created = fmt.Sprintf("%d hours ago", hours)
			}

		case minutes > 0:
			if minutes == 1 {
				created = "1 minute ago"
			} else {
				created = fmt.Sprintf("%d minutes ago", minutes)
			}

		default:
			created = "Just now"
		}
	}

	return created
}

func (d Deployment) OSRelease() map[string]string {
	values := make(map[string]string)

	file, err := os.Open(d.Path() + "/etc/os-release")
	if err != nil {
		return values
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		tmp := strings.Split(scanner.Text(), "=")

		values[tmp[0]] = tmp[1]
	}

	return values
}
