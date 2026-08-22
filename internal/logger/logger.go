package logger

// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) Lewis Cook <hi@lcook.net>

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"
)

type Logger struct {
	prefix         string
	stdout, stderr io.Writer
}

var logger *Logger

func init() {
	logger = &Logger{
		prefix: "*",
		stdout: os.Stdout,
		stderr: os.Stderr,
	}
}

func Info(msg string) {
	fmt.Fprintf(
		logger.stdout,
		"%s %s\n",
		color.New(color.FgGreen).Sprint(logger.prefix),
		msg,
	)
}

func Warn(msg string) {
	fmt.Fprintf(
		logger.stdout,
		"%s %s\n",
		color.New(color.FgYellow).Sprint("warn:"),
		msg,
	)
}

func Error(msg string) {
	fmt.Fprintf(
		logger.stderr,
		"%s %s\n",
		color.New(color.FgRed).Sprint("error:"),
		msg,
	)
	os.Exit(1)
}

func Input(msg string) string {
	fmt.Fprintf(
		logger.stdout,
		"%s %s",
		color.New(color.FgGreen).Sprint(logger.prefix),
		msg,
	)

	input := bufio.NewScanner(os.Stdin)
	input.Scan()

	return input.Text()
}

func Run(desc string, cmd []string) error {
	if desc != "" {
		Info(desc)
	}

	fmt.Fprintf(
		logger.stdout,
		"%s\n",
		color.New(color.FgYellow).Sprint(" "+strings.Join(cmd, " ")),
	)

	execution := exec.Command(cmd[0], cmd[1:]...)

	out, err := execution.CombinedOutput()
	fmt.Fprint(logger.stdout, string(out))

	return err
}

func Infof(
	format string,
	args ...any,
) {
	Info(fmt.Sprintf(format, args...))
}

func Warnf(
	format string,
	args ...any,
) {
	Warn(fmt.Sprintf(format, args...))
}

func Errorf(
	format string,
	args ...any,
) {
	Error(fmt.Sprintf(format, args...))
}

func Inputf(
	format string,
	args ...any,
) string {
	return Input(fmt.Sprintf(format, args...))
}

func Runf(cmd []string, format string, args ...any) error {
	return Run(fmt.Sprintf(format, args...), cmd)
}
