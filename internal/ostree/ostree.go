package ostree

// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) Lewis Cook <hi@lcook.net>

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/lcook/yorha/internal/logger"
	"github.com/lcook/yorha/internal/podman"
	"github.com/lcook/yorha/internal/util"
)

const (
	DefaultStateroot = "yorha"
	DefaultBranch    = "yorha/x86_64/standard"
)

type Options struct {
	SysRoot  string
	SysSetup string
	SysTree  string

	Image      string
	PartLabels map[string]string

	Interactive bool
	ForceUpdate bool

	Log *logger.Logger
}

func Environment() bool {
	if _, err := os.Stat("/run/ostree-booted"); err != nil {
		return false
	}

	return true
}

func CreateRootFilesystem(opt *Options) {
	if _, err := os.Stat(opt.SysTree); err == nil {
		opt.Log.Infof("Removing existing setup directory at %s", opt.SysTree)

		if err := os.RemoveAll(opt.SysTree); err != nil {
			opt.Log.Error(err.Error())
		}
	}

	opt.Log.Infof("Creating setup directory at %s", opt.SysTree)

	if err := os.MkdirAll(opt.SysTree, 0o755); err != nil {
		opt.Log.Error(err.Error())
	}

	podman, err := podman.NewClient(podman.RootfullContext)
	if err != nil {
		opt.Log.Error(err.Error())
	}

	var (
		output = opt.SysSetup + "/rootfs.tar"
		image  = opt.Image
	)

	if opt.Interactive {
		image = opt.Log.Inputf(
			"Enter container image for deployment [%s]: ",
			opt.Image,
		)

		if image == "" {
			image = opt.Image
		}
	}

	if !podman.HasLocalImage(image) {
		opt.Log.Infof("Container image %s not found in local storage", image)

		if err := podman.PullImage(image); err != nil {
			opt.Log.Errorf(
				"Failed to pull container image %s: %s",
				image,
				err.Error(),
			)
		}
	} else if Environment() && !opt.ForceUpdate {
		opt.Log.Infof("Comparing local and remote digests for image %s", image)

		inspect, err := podman.GetImage(image)
		if err != nil {
			opt.Log.Error(err.Error())
		}

		remote, err := podman.GetRemoteImage("//" + image)
		if err != nil {
			opt.Log.Error(err.Error())
		}

		if inspect.ID == strings.TrimPrefix(remote.Digest.String(), "sha256:") {
			opt.Log.Errorf(
				"No update available (local:%s, remote:%s)",
				inspect.ID[0:11],
				strings.TrimPrefix(remote.Digest.String(), "sha256:")[0:11],
			)
		}

		opt.Log.Infof(
			"Container image update available (local:%s, remote:%s)",
			inspect.ID[0:11],
			strings.TrimPrefix(remote.Digest.String(), "sha256:")[0:11],
		)

		if err := podman.PullImage(image); err != nil {
			opt.Log.Errorf(
				"Failed to pull latest image %s: %s",
				image,
				err.Error(),
			)
		}

		if err := podman.RemoveLocalImage(inspect.ID); err != nil {
			opt.Log.Errorf(
				"Failed to remove previous image %s: %s",
				inspect.ID[0:11],
				err.Error(),
			)
		}

		opt.Log.Infof(
			"Removed old local container image %s:%s",
			image,
			inspect.ID[0:11],
		)
	}

	opt.Log.Info(
		"Preparing OSTree filesystem from container image",
	)

	handle, err := util.GetFileDescriptor(output)
	if err != nil {
		opt.Log.Error(err.Error())
	}

	opt.Log.Infof("Exporting container image %s to archive %s",
		image,
		output,
	)

	if err := podman.ExportContainer(image, handle); err != nil {
		opt.Log.Error(err.Error())
	}

	opt.Log.Run("Extracting container root filesystem archive", []string{
		"tar",
		"xf",
		output,
		"-C",
		opt.SysTree,
	})
}

func CreateLayout(opt *Options) {
	os.Create(opt.SysTree + "/etc/machine-id")

	os.Rename(opt.SysTree+"/etc", opt.SysTree+"/usr/etc")

	os.RemoveAll(opt.SysTree + "/home")
	os.Symlink("/var/home", opt.SysTree+"/home")

	os.RemoveAll(opt.SysTree + "/mnt")
	os.Symlink("/var/mnt", opt.SysTree+"/mnt")

	os.RemoveAll(opt.SysTree + "/root")
	os.Symlink("/var/roothome", opt.SysTree+"/root")

	os.RemoveAll(opt.SysTree + "/srv")
	os.Symlink("/var/srv", opt.SysTree+"/srv")

	os.MkdirAll(opt.SysTree+"/sysroot", 0o755)
	os.Symlink("/sysroot/ostree", opt.SysTree+"/ostree")

	os.RemoveAll(opt.SysTree + "/usr/local")
	os.Symlink("/var/usrlocal", opt.SysTree+"/usr/local")

	opt.Log.Infof("Created OSTree filesystem layout at %s",
		opt.SysTree,
	)

	opt.Log.Info("Writing systemd-tmpfiles(8) configuration")

	os.WriteFile(
		opt.SysTree+"/usr/lib/tmpfiles.d/ostree-0-integration.conf",
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
		opt.SysTree+"/var/lib/pacman",
		opt.SysTree+"/usr/lib/pacman",
	)

	content, _ := os.ReadFile(opt.SysTree + "/usr/etc/pacman.conf")
	lines := strings.Split(string(content), "\n")

	for i, line := range lines {
		if strings.HasPrefix(line, "#DBPath") {
			lines[i] = "DBPath = /usr/lib/pacman/"
		} else if strings.HasPrefix(line, "#IgnoreGroup") {
			lines[i] = "IgnoreGroup = modified"
		}
	}

	os.WriteFile(
		opt.SysTree+"/usr/etc/pacman.conf",
		[]byte(strings.Join(lines, "\n")),
		0o644,
	)

	matches, _ := filepath.Glob(
		filepath.Join(opt.SysTree, "var", "*"),
	)
	for _, m := range matches {
		os.RemoveAll(m)
	}

	opt.Log.Run("", []string{"chmod u-s", opt.SysTree + "/usr/bin/newuidmap"})
	opt.Log.Run("", []string{"chmod u-s", opt.SysTree + "/usr/bin/newgidmap"})

	opt.Log.Run(
		"Restoring user namespace capability on newuidmap",
		[]string{
			"setcap",
			"cap_setuid+eip",
			opt.SysTree + "/usr/bin/newuidmap",
		},
	)

	opt.Log.Run(
		"Restoring group namespace capability on newgidmap",
		[]string{
			"setcap",
			"cap_setgid+eip",
			opt.SysTree + "/usr/bin/newgidmap",
		})
}

func DeployImage(opt *Options) {
	opt.Log.Runf(
		[]string{
			"ostree",
			"commit",
			"--repo=" + filepath.Join(opt.SysRoot, "ostree", "repo"),
			"--branch=" + DefaultBranch,
			"--tree=dir=" + opt.SysTree,
		},
		"Committing new root filesystem to OSTree branch %s from directory %s",
		DefaultBranch,
		opt.SysTree,
	)

	var (
		cmd = []string{
			"ostree",
			"admin",
			"deploy",
			"--sysroot=" + opt.SysRoot,
		}
		kargs = []string{
			"--karg-none",
			"--karg=root=LABEL=SYS_ROOT",
			"--karg=rw",
		}
	)

	kargFile := fmt.Sprintf("/etc/%s/kargs", DefaultStateroot)
	if !Environment() {
		kargFile = path.Join(opt.SysTree, "usr", kargFile)
	}

	if _, err := os.Stat(kargFile); err == nil {
		opt.Log.Infof("Applying kernel arguments from %s", kargFile)

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

	opt.Log.Run("Deploying OSTree revision", cmd)

	if Environment() {
		opt.Log.Run("Regenerating GRUB configuration", []string{
			"grub-mkconfig",
			"-o",
			"/boot/efi/EFI/grub/grub.cfg",
		})
	}
}
