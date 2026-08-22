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
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"text/tabwriter"

	log "github.com/lcook/yorha/internal/logger"
	"github.com/lcook/yorha/internal/podman"
	"github.com/lcook/yorha/internal/util"
)

type Manager struct{ Config }

func New(config Config) *Manager { return &Manager{config} }

func (m *Manager) CreateRootFilesystem() {
	if _, err := os.Stat(m.SysTree); err == nil {
		log.Infof("Removing existing setup directory at %s", m.SysTree)

		if err := os.RemoveAll(m.SysTree); err != nil {
			log.Error(err.Error())
		}
	}

	log.Infof("Creating setup directory at %s", m.SysTree)

	if err := os.MkdirAll(m.SysTree, 0o755); err != nil {
		log.Error(err.Error())
	}

	podman, err := podman.NewClient(podman.RootfullContext)
	if err != nil {
		log.Error(err.Error())
	}

	var (
		output = m.SysSetup + "/rootfs.tar"
		image  = m.Image
	)

	if m.Interactive {
		image = log.Inputf(
			"Enter container image for deployment [%s]: ",
			m.Image,
		)

		if image == "" {
			image = m.Image
		}
	}

	var local bool
	if strings.Split(image, "/")[0] == "localhost" {
		local = true
	}

	if !podman.HasLocalImage(image) {
		log.Infof("Container image %s not found in local storage", image)

		if err := podman.PullImage(image); err != nil {
			log.Errorf(
				"Failed to pull container image %s: %s",
				image,
				err.Error(),
			)
		}
	} else if !local && !m.ForceUpdate {
		log.Infof("Comparing local and remote digests for image %s", image)

		inspect, err := podman.GetImage(image)
		if err != nil {
			log.Error(err.Error())
		}

		remote, err := podman.GetRemoteImage("//" + image)
		if err != nil {
			log.Error(err.Error())
		}

		if inspect.ID == strings.TrimPrefix(remote.Digest.String(), "sha256:") {
			log.Errorf(
				"No update available (local:%s, remote:%s)",
				inspect.ID[0:11],
				strings.TrimPrefix(remote.Digest.String(), "sha256:")[0:11],
			)
		}

		log.Infof(
			"Container image update available (local:%s, remote:%s)",
			inspect.ID[0:11],
			strings.TrimPrefix(remote.Digest.String(), "sha256:")[0:11],
		)

		if err := podman.PullImage(image); err != nil {
			log.Errorf(
				"Failed to pull latest image %s: %s",
				image,
				err.Error(),
			)
		}

		if err := podman.RemoveLocalImage(inspect.ID); err != nil {
			log.Errorf(
				"Failed to remove previous image %s: %s",
				inspect.ID[0:11],
				err.Error(),
			)
		}

		log.Infof(
			"Removed old local container image %s:%s",
			image,
			inspect.ID[0:11],
		)
	}

	log.Info(
		"Preparing OSTree filesystem from container image",
	)

	handle, err := util.GetFileDescriptor(output)
	if err != nil {
		log.Error(err.Error())
	}

	log.Infof(
		"Exporting container image %s to archive %s",
		image,
		output,
	)

	if err := podman.ExportContainer(image, handle); err != nil {
		log.Error(err.Error())
	}

	log.Run("Extracting container root filesystem archive", []string{
		"tar",
		"xf",
		output,
		"-C",
		m.SysTree,
	})
}

func (m *Manager) CreateLayout() {
	os.Create(m.SysTree + "/etc/machine-id")

	os.Rename(m.SysTree+"/etc", m.SysTree+"/usr/etc")

	os.RemoveAll(m.SysTree + "/home")
	os.Symlink("/var/home", m.SysTree+"/home")

	os.RemoveAll(m.SysTree + "/mnt")
	os.Symlink("/var/mnt", m.SysTree+"/mnt")

	os.RemoveAll(m.SysTree + "/root")
	os.Symlink("/var/roothome", m.SysTree+"/root")

	os.RemoveAll(m.SysTree + "/srv")
	os.Symlink("/var/srv", m.SysTree+"/srv")

	os.MkdirAll(m.SysTree+"/sysroot", 0o755)
	os.Symlink("/sysroot/ostree", m.SysTree+"/ostree")

	os.RemoveAll(m.SysTree + "/usr/local")
	os.Symlink("/var/usrlocal", m.SysTree+"/usr/local")

	log.Infof(
		"Created OSTree filesystem layout at %s",
		m.SysTree,
	)

	log.Info("Writing systemd-tmpfiles(8) configuration")

	os.WriteFile(
		m.SysTree+"/usr/lib/tmpfiles.d/ostree-0-integration.conf",
		[]byte(`d /var/home 0755 root root -
d /var/lib 0755 root root -
d /var/log/journal 0755 root root -
d /var/mnt 0755 root root -
d /var/opt 0755 root root -
d /var/roothome 0700 root root -
d /var/srv 0755 root root -
d /var/usrlocal 0755 root root -
d /var/usrlocal/bin 0755 root root -
d /var/usrlocal/etc 0755 root root -
d /var/usrlocal/games 0755 root root -
d /var/usrlocal/include 0755 root root -
d /var/usrlocal/lib 0755 root root -
d /var/usrlocal/man 0755 root root -
d /var/usrlocal/sbin 0755 root root -
d /var/usrlocal/share 0755 root root -
d /var/usrlocal/src 0755 root root -
d /run/media 0755 root root -`),
		0o755,
	)

	os.Rename(
		m.SysTree+"/var/lib/pacman",
		m.SysTree+"/usr/lib/pacman",
	)

	content, _ := os.ReadFile(m.SysTree + "/usr/etc/pacman.conf")
	lines := strings.Split(string(content), "\n")

	for i, line := range lines {
		if strings.HasPrefix(line, "#DBPath") {
			lines[i] = "DBPath = /usr/lib/pacman/"
		} else if strings.HasPrefix(line, "#IgnoreGroup") {
			lines[i] = "IgnoreGroup = modified"
		}
	}

	os.WriteFile(
		m.SysTree+"/usr/etc/pacman.conf",
		[]byte(strings.Join(lines, "\n")),
		0o644,
	)

	matches, _ := filepath.Glob(
		filepath.Join(m.SysTree, "var", "*"),
	)
	for _, m := range matches {
		os.RemoveAll(m)
	}

	log.Run("", []string{"chmod u-s", m.SysTree + "/usr/bin/newuidmap"})
	log.Run("", []string{"chmod u-s", m.SysTree + "/usr/bin/newgidmap"})

	log.Run(
		"Restoring user namespace capability on newuidmap",
		[]string{
			"setcap",
			"cap_setuid+eip",
			m.SysTree + "/usr/bin/newuidmap",
		},
	)

	log.Run(
		"Restoring group namespace capability on newgidmap",
		[]string{
			"setcap",
			"cap_setgid+eip",
			m.SysTree + "/usr/bin/newgidmap",
		},
	)
}

func (m *Manager) DeployImage() {
	log.Runf(
		[]string{
			"ostree",
			"commit",
			"--repo=" + filepath.Join(m.SysRoot, "ostree", "repo"),
			"--branch=" + DefaultBranch,
			"--tree=dir=" + m.SysTree,
		},
		"Committing new root filesystem to OSTree branch %s from directory %s",
		DefaultBranch,
		m.SysTree,
	)

	var (
		cmd = []string{
			"ostree",
			"admin",
			"deploy",
			"--sysroot=" + m.SysRoot,
		}
		kargs = []string{
			"--karg-none",
			"--karg=root=LABEL=SYS_ROOT",
			"--karg=rw",
		}
	)

	kargFile := fmt.Sprintf("/etc/%s/kargs", DefaultStateroot)
	if !Environment() {
		kargFile = path.Join(m.SysTree, "usr", kargFile)
	}

	if _, err := os.Stat(kargFile); err == nil {
		log.Infof("Applying kernel arguments from %s", kargFile)

		file, _ := os.Open(kargFile)

		scanner := bufio.NewScanner(file)
		scanner.Split(bufio.ScanLines)

		for scanner.Scan() {
			kargs = append(kargs, fmt.Sprintf("--karg=%s", scanner.Text()))
		}

		file.Close()
	}

	cmd = append(cmd, kargs...)
	cmd = append(cmd, "--os="+DefaultStateroot)
	cmd = append(cmd, "--retain")
	cmd = append(cmd, DefaultBranch)

	log.Run("Deploying OSTree revision", cmd)

	if Environment() {
		log.Run("Regenerating GRUB configuration", []string{
			"grub-mkconfig",
			"-o",
			"/boot/efi/EFI/grub/grub.cfg",
		})
	}
}

func (m *Manager) PrintDeployments() {
	deployments, err := m.GetDeployments()
	if err != nil {
		log.Error(err.Error())
	}

	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(writer, "\tIMAGE\tDEPLOYMENT\tVERSION\tCREATED\tSTATUS")

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

		fmt.Fprintf(
			writer, "%2d\t%s\t%s\t%s\t%s\t%s\n",
			deployment.Index,
			release["IMAGE"],
			deployment.Checksum[0:11],
			release["VERSION_ID"],
			deployment.Created(),
			strings.Join(status, " "),
		)
	}

	writer.Flush()
}

func (m *Manager) GetDeployments() ([]Deployment, error) {
	cmd := exec.Command(
		"ostree",
		"admin",
		"status",
		"--sysroot="+m.SysRoot,
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

func (m *Manager) SwitchDeployment(index int) error {
	_, err := exec.Command(
		"ostree",
		"admin",
		"set-default",
		strconv.Itoa(index),
	).Output()

	return err
}
