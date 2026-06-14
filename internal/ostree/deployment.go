package ostree

// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) Lewis Cook <hi@lcook.net>

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"text/tabwriter"
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

	if val, ok := release["VERSION"]; ok {
		version, err := time.ParseInLocation("20060102.1504", val, time.Local)
		if err != nil {
			return created
		}

		var (
			duration = time.Since(version)

			days    = int(duration.Hours() / 24)
			hours   = int(duration.Hours()) % 24
			minutes = int(duration.Minutes()) % 60
		)

		if days > 0 {
			if days == 1 {
				return fmt.Sprintf("%d day ago", days)
			} else {
				created = fmt.Sprintf("%d days ago", days)
			}
		} else if hours > 0 {
			if hours == 1 {
				created = fmt.Sprintf("%d hour ago", hours)
			} else {
				created = fmt.Sprintf("%d hours ago", hours)
			}
		} else if minutes > 0 {
			if minutes == 1 {
				created = fmt.Sprintf("%d minute ago", minutes)
			} else {
				created = fmt.Sprintf("%d minutes ago", minutes)
			}
		} else {
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

func PrintDeployments(opt *Options) {
	deployments, err := GetDeployments(opt)
	if err != nil {
		opt.Log.Error(err.Error())
	}

	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(writer, "\tIMAGE\tCHECKSUM\tVERSION\tCREATED\tSTATUS")

	for _, deployment := range deployments {
		var status []string

		switch true {
		case deployment.Staged:
			status = append(status, "staged")
		case deployment.Pending:
			status = append(status, "pending")
		case deployment.Rollback:
			status = append(status, "rollback")
		case deployment.Pinned:
			status = append(status, "pinned")
		case deployment.Booted:
			status = append(status, "booted")
		}

		release := deployment.OSRelease()

		fmt.Fprintf(writer, "%2d\t%s\t%s\t%s\t%s\t%s\n",
			deployment.Index,
			release["IMAGE"],
			deployment.Checksum[0:11],
			release["VERSION"],
			deployment.Created(),
			strings.Join(status, " "),
		)
	}

	writer.Flush()
}

func GetDeployments(opt *Options) ([]Deployment, error) {
	cmd := exec.Command(
		"ostree",
		"admin",
		"status",
		"--sysroot="+opt.SysRoot,
		"--json",
	)

	out, err := cmd.Output()
	if err != nil {
		return []Deployment{}, err
	}

	var w struct {
		Deployments []Deployment `json:"deployments"`
	}

	err = json.Unmarshal(out, &w)
	if err != nil || len(w.Deployments) == 0 {
		return []Deployment{}, errors.New("no deployments found")
	}

	return w.Deployments, nil
}

func SwitchDeployment(opt *Options, index int) error {
	_, err := exec.Command(
		"ostree",
		"admin",
		"set-default",
		strconv.Itoa(index),
	).Output()

	return err
}
