package config

// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) Lewis Cook <hi@lcook.net>

import (
	"os"

	"gopkg.in/yaml.v3"
)

func FromFile[T any](path string) (T, error) {
	var settings T

	contents, err := os.ReadFile(path)
	if err != nil {
		return settings, err
	}

	err = yaml.Unmarshal(contents, &settings)
	if err != nil {
		return settings, err
	}

	return settings, nil
}
