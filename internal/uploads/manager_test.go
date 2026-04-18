package uploads

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"mime/multipart"
	"net/textproto"
	"os"
	"path/filepath"
	"testing"

	inerrors "github.com/mikhail5545/wasmforge/internal/errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestManagerFromMultipartFileCalculatesHashFromUploadedContent(t *testing.T) {
	mgr, pluginDir := newTestManager(t)
	payload := []byte("wasm-payload-v1")
	fileHeader := newMultipartFileHeader(t, "plugin.wasm", payload, nil)

	hash, err := mgr.FromMultipartFile(fileHeader, "plugin.wasm", PluginUpload)
	require.NoError(t, err)

	expectedHash := sha256.Sum256(payload)
	require.Equal(t, hex.EncodeToString(expectedHash[:]), hash)

	storedData, err := os.ReadFile(filepath.Join(pluginDir, "plugin.wasm"))
	require.NoError(t, err)
	require.Equal(t, payload, storedData)
}

func TestManagerRejectsPathTraversalFilenames(t *testing.T) {
	mgr, _ := newTestManager(t)

	tests := []struct {
		name     string
		filename string
	}{
		{name: "parent traversal", filename: "../escape.wasm"},
		{name: "nested path", filename: "nested/escape.wasm"},
		{name: "absolute path", filename: "/escape.wasm"},
		{name: "windows traversal", filename: `..\escape.wasm`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := mgr.FromBytes([]byte("x"), tc.filename, PluginUpload)
			require.Error(t, err)
			require.Contains(t, err.Error(), "invalid filename")
		})
	}
}

func TestManagerSizeLimitErrorsAreConsistent(t *testing.T) {
	mgr, _ := newTestManager(t)

	oversized := make([]byte, maxUploadSizeBytes+1)
	_, err := mgr.FromBytes(oversized, "large.wasm", PluginUpload)
	require.Error(t, err)
	require.True(t, errors.Is(err, inerrors.ErrSizeLimitExceeded))
	require.Contains(t, err.Error(), maxUploadSizeLabel)

	claimedSize := int64(maxUploadSizeBytes + 1)
	fileHeader := newMultipartFileHeader(t, "large-multipart.wasm", []byte("small"), &claimedSize)
	_, err = mgr.FromMultipartFile(fileHeader, "large-multipart.wasm", PluginUpload)
	require.Error(t, err)
	require.True(t, errors.Is(err, inerrors.ErrSizeLimitExceeded))
	require.Contains(t, err.Error(), maxUploadSizeLabel)
}

func newTestManager(t *testing.T) (*manager, string) {
	t.Helper()

	pluginDir, err := os.MkdirTemp(".", "plugin-upload-test-*")
	require.NoError(t, err)
	certDir, err := os.MkdirTemp(".", "cert-upload-test-*")
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = os.RemoveAll(pluginDir)
		_ = os.RemoveAll(certDir)
	})

	mgr := New(pluginDir, certDir, zap.NewNop())
	typed, ok := mgr.(*manager)
	require.True(t, ok)
	return typed, pluginDir
}

func newMultipartFileHeader(t *testing.T, filename string, payload []byte, overrideSize *int64) *multipart.FileHeader {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="wasm_file"; filename="%s"`, filename))
	header.Set("Content-Type", "application/wasm")

	part, err := writer.CreatePart(header)
	require.NoError(t, err)
	_, err = part.Write(payload)
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	reader := multipart.NewReader(bytes.NewReader(body.Bytes()), writer.Boundary())
	form, err := reader.ReadForm(int64(body.Len() + 1024))
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = form.RemoveAll()
	})

	files := form.File["wasm_file"]
	require.Len(t, files, 1)
	if overrideSize != nil {
		files[0].Size = *overrideSize
	}
	return files[0]
}
