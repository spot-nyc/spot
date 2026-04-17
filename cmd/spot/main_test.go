package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/spot-nyc/spot"
)

func TestRootCommand_Structure(t *testing.T) {
	cmd := newRootCmd()

	assert.Equal(t, "spot", cmd.Use)
	assert.NotEmpty(t, cmd.Short, "root command should have a short description")
	assert.NotEmpty(t, cmd.Long, "root command should have a long description")
	assert.Equal(t, spot.Version, cmd.Version, "root command version should track spot.Version")
}

func TestRootCommand_VersionFlag(t *testing.T) {
	cmd := newRootCmd()

	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--version"})

	err := cmd.Execute()
	assert.NoError(t, err)

	assert.Contains(t, out.String(), spot.Version,
		"version output should contain spot.Version (%q)", spot.Version)
}

func TestRootCommand_HelpFlag(t *testing.T) {
	cmd := newRootCmd()

	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	assert.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, "spot", "help output should name the command")
	assert.Contains(t, output, "Usage:", "help output should have a Usage section")
}
