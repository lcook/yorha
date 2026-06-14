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

func New() *Logger {
	return &Logger{
		prefix: "*",
		stdout: os.Stdout,
		stderr: os.Stderr,
	}
}

func (l *Logger) Info(msg string) {
	fmt.Fprintf(
		l.stdout,
		"%s %s\n",
		color.New(color.FgGreen).Sprint(l.prefix),
		msg,
	)
}

func (l *Logger) Warn(msg string) {
	fmt.Fprintf(
		l.stdout,
		"%s %s\n",
		color.New(color.FgYellow).Sprint("warn:"),
		msg,
	)
}

func (l *Logger) Error(msg string) {
	fmt.Fprintf(
		l.stderr,
		"%s %s\n",
		color.New(color.FgRed).Sprint("error:"),
		msg,
	)
	os.Exit(1)
}

func (l *Logger) Input(msg string) string {
	fmt.Fprintf(
		l.stdout,
		"%s %s",
		color.New(color.FgGreen).Sprint(l.prefix),
		msg,
	)

	input := bufio.NewScanner(os.Stdin)
	input.Scan()

	return input.Text()
}

func (l *Logger) Run(desc string, cmd []string) error {
	if desc != "" {
		l.Info(desc)
	}

	fmt.Fprintf(
		l.stdout,
		"%s\n",
		color.New(color.FgYellow).Sprint(" "+strings.Join(cmd, " ")),
	)

	execution := exec.Command(cmd[0], cmd[1:]...)

	out, err := execution.CombinedOutput()
	fmt.Fprint(l.stdout, string(out))

	return err
}

func (l *Logger) Infof(
	format string,
	args ...any,
) {
	l.Info(fmt.Sprintf(format, args...))
}

func (l *Logger) Warnf(
	format string,
	args ...any,
) {
	l.Warn(fmt.Sprintf(format, args...))
}

func (l *Logger) Errorf(
	format string,
	args ...any,
) {
	l.Error(fmt.Sprintf(format, args...))
}

func (l *Logger) Inputf(format string, args ...any) string {
	return l.Input(fmt.Sprintf(format, args...))
}

func (l *Logger) Runf(cmd []string, format string, args ...any) error {
	return l.Run(fmt.Sprintf(format, args...), cmd)
}
