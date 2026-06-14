package podman

// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) Lewis Cook <hi@lcook.net>

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"go.podman.io/buildah/define"
	"go.podman.io/image/v5/docker"
	imageTypes "go.podman.io/image/v5/types"
	"go.podman.io/podman/v6/pkg/bindings"
	"go.podman.io/podman/v6/pkg/bindings/containers"
	"go.podman.io/podman/v6/pkg/bindings/images"
	"go.podman.io/podman/v6/pkg/domain/entities/types"
	"go.podman.io/podman/v6/pkg/specgen"

	"github.com/lcook/yorha/internal/config"
	"github.com/lcook/yorha/internal/util"
)

type ConnectionType int

const (
	RootfullContext ConnectionType = iota
	RootlessContext
)

type PodmanClient struct {
	ctx context.Context
}

func NewClient(ct ConnectionType) (*PodmanClient, error) {
	var socket string

	switch ct {
	case RootfullContext:
		socket = "unix:///run/podman/podman.sock"
	case RootlessContext:
		socket = fmt.Sprintf(
			"unix:///run/user/%d/podman/podman.sock",
			os.Getuid(),
		)
	default:
		return nil, fmt.Errorf("invalid connection type")
	}

	conn, err := bindings.NewConnection(context.Background(), socket)
	if err != nil {
		return nil, err
	}

	return &PodmanClient{ctx: conn}, nil
}

func (p *PodmanClient) BuildContainer(container config.Container) error {
	file, err := os.CreateTemp("", "")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())

	err = util.GenerateContainerfile(file, container)
	if err != nil {
		return err
	}

	_, err = images.Build(
		p.ctx,
		[]string{file.Name()},
		types.BuildOptions{BuildOptions: define.BuildOptions{
			ContextDirectory: filepath.Dir("."),
			Output:           container.Image,
		}},
	)
	if err != nil {
		return err
	}

	return nil
}

func (p *PodmanClient) CreateContainer(
	image string,
) (string, error) {
	spec := specgen.NewSpecGenerator(image, false)
	spec.Command = []string{"sh"}

	container, err := containers.CreateWithSpec(p.ctx, spec, nil)
	if err != nil {
		return "", err
	}

	return container.ID, nil
}

func (p *PodmanClient) ExportContainer(
	image string,
	handle io.Writer,
) error {
	id, err := p.CreateContainer(image)
	if err != nil {
		return err
	}

	err = containers.Export(p.ctx, id, handle, nil)
	if err != nil {
		return err
	}

	_, err = containers.Remove(p.ctx, id, nil)
	if err != nil {
		return err
	}

	return nil
}

func (p *PodmanClient) GetImages() ([]*types.ImageSummary, error) {
	local, err := images.List(
		p.ctx,
		&images.ListOptions{All: new(false)},
	)
	if err != nil {
		return nil, err
	}

	return local, nil
}

func (p *PodmanClient) HasLocalImage(image string) bool {
	if image == "" {
		return false
	}

	local, err := p.GetImages()
	if err != nil || len(local) == 0 {
		return false
	}

	image = strings.TrimSpace(image)
	if strings.Contains(image, ":") {
		parts := strings.Split(image, ":")
		if len(parts) != 2 || parts[1] == "" {
			return false
		}
	} else {
		image = image + ":latest"
	}

	var found bool

	for _, img := range local {
		if len(img.Names) == 0 {
			continue
		}

		for _, name := range img.Names {
			if name == image {
				found = true
			}
		}
	}

	return found
}

func (p *PodmanClient) PullImage(image string) error {
	_, err := images.Pull(p.ctx, image, &images.PullOptions{})

	return err
}

func (p *PodmanClient) GetImage(
	image string,
) (*types.ImageInspectReport, error) {
	return images.GetImage(p.ctx, image, &images.GetOptions{})
}

func (p *PodmanClient) GetRemoteImage(
	image string,
) (imageTypes.BlobInfo, error) {
	ref, err := docker.ParseReference(image)
	if err != nil {
		return imageTypes.BlobInfo{}, err
	}

	img, err := ref.NewImage(p.ctx, &imageTypes.SystemContext{})
	if err != nil {
		return imageTypes.BlobInfo{}, err
	}
	defer img.Close()

	return img.ConfigInfo(), err
}

func (p *PodmanClient) RemoveLocalImage(
	image string,
) error {
	_, err := images.Remove(
		p.ctx,
		[]string{image},
		&images.RemoveOptions{Force: new(true)},
	)
	if err != nil {
		return err[0]
	}

	return nil
}
