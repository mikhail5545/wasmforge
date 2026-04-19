package plugins

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateRequestValidateVersionConstraint(t *testing.T) {
	req := CreateRequest{
		RouteID:           "01983a0a-74d6-76af-a377-56f8a6f14512",
		PluginID:          "01983a0a-74d6-76af-a377-56f8a6f14513",
		VersionConstraint: "^1.2",
		ExecutionOrder:    1,
	}
	require.NoError(t, req.Validate())

	req.VersionConstraint = "invalid-constraint"
	require.Error(t, req.Validate())
}

func TestUpdateRequestValidateVersionConstraint(t *testing.T) {
	req := UpdateRequest{
		ID:                "01983a0a-74d6-76af-a377-56f8a6f14514",
		VersionConstraint: ptr("^2.0"),
	}
	require.NoError(t, req.Validate())

	req.VersionConstraint = ptr("nope")
	require.Error(t, req.Validate())
}

func ptr[T any](v T) *T {
	return &v
}
