package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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

func TestRootCommand_JSONFlagParsing(t *testing.T) {
	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--json"})

	// With --json but no subcommand, root's RunE still just runs help — but
	// the flag parse should succeed and the Bool value should be true.
	require.NoError(t, cmd.Execute())

	flag := cmd.PersistentFlags().Lookup("json")
	require.NotNil(t, flag, "persistent --json flag should be registered")
	assert.Equal(t, "true", flag.Value.String())
}

func TestRootCommand_JFlagIsShortForm(t *testing.T) {
	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"-j"})

	require.NoError(t, cmd.Execute())
	assert.Equal(t, "true", cmd.PersistentFlags().Lookup("json").Value.String())
}
