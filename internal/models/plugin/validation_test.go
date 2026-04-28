package plugin

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateRequestValidateVersion(t *testing.T) {
	req := CreateRequest{
		Name:     "plugin_alpha",
		Version:  "1.2.3-beta.1+build.5",
		Filename: "plugin_alpha.wasm",
	}
	require.NoError(t, req.Validate())

	req.Version = "1.2"
	require.Error(t, req.Validate())
}

func TestListRequestValidateVersions(t *testing.T) {
	req := ListRequest{
		Versions:       []string{"1.0.0", "2.1.0-rc.1"},
		OrderField:     OrderFieldVersion,
		OrderDirection: "asc",
		PageSize:       10,
	}
	require.NoError(t, req.Validate())

	req.Versions = []string{"invalid-version"}
	require.Error(t, req.Validate())
}
