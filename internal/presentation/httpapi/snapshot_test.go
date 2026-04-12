package httpapi_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/bradleyjkemp/cupaloy/v2"
	"github.com/stretchr/testify/require"
)

var snapshotter = cupaloy.New(cupaloy.SnapshotFileExtension(".json"))

func snapshotJSON(t *testing.T, rawJSON string) {
	t.Helper()
	if rawJSON == "" {
		snapshotter.SnapshotT(t, rawJSON)
		return
	}
	var buf bytes.Buffer
	require.NoError(t, json.Indent(&buf, []byte(rawJSON), "", "  "))
	snapshotter.SnapshotT(t, buf.String())
}
