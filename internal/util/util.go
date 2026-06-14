package util

// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) Lewis Cook <hi@lcook.net>

import (
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/lcook/yorha/internal/config"
)

func GenerateContainerfile(handle io.Writer, container config.Container) error {
	funcs := template.FuncMap{
		"ServiceCommands": ServiceCommands,
		"Packages":        Packages,
		"Date":            func() time.Time { return time.Now() },
	}

	return template.Must(template.New("t").Funcs(funcs).Parse(container.Template)).
		Execute(handle, container)
}

func ServiceCommands(action string, services []string) string {
	if len(services) == 0 {
		return ""
	}

	var lines []string
	for _, svc := range services {
		lines = append(
			lines,
			fmt.Sprintf("RUN systemctl %s %s.service", action, svc),
		)
	}

	return strings.Join(lines, "\n")
}

func Packages(packages []string) string {
	if len(packages) == 0 {
		return ""
	}

	return " \\\n    " + strings.Join(packages, " \\\n    ")
}

func GetFileDescriptor(path string) (*os.File, error) {
	if path == "" {
		return os.Stdout, nil
	}

	handle, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0o600)
	if err != nil {
		return nil, err
	}

	return handle, nil
}
