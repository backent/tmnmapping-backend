package buildingimage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSave_WritesUnderTheBuildingsOwnDirectory(t *testing.T) {
	storage := NewStorage(t.TempDir())

	relative, err := storage.Save(42, "front", ".jpg", []byte("not really a jpeg"))

	assert.NoError(t, err)
	assert.True(t, strings.HasPrefix(relative, "42"+string(os.PathSeparator)), relative)
	assert.True(t, strings.HasSuffix(relative, ".jpg"), relative)

	file, err := storage.Open(relative)
	assert.NoError(t, err)
	defer file.Close()
}

// Replacing a photo must change its URL. Reusing <building>/<slot>.jpg would leave a
// browser showing the old picture from cache with nothing to invalidate it.
func TestSave_TwoUploadsToOneSlotGetDifferentNames(t *testing.T) {
	storage := NewStorage(t.TempDir())

	first, err := storage.Save(42, "front", ".jpg", []byte("one"))
	assert.NoError(t, err)

	second, err := storage.Save(42, "front", ".jpg", []byte("two"))
	assert.NoError(t, err)

	assert.NotEqual(t, first, second)
}

// The paths come from our own database, but a traversal bug here would serve
// arbitrary files off the volume, so it is checked rather than assumed.
func TestOpen_RefusesToClimbOutOfTheRoot(t *testing.T) {
	root := t.TempDir()
	secret := filepath.Join(filepath.Dir(root), "secret.txt")
	assert.NoError(t, os.WriteFile(secret, []byte("not yours"), 0o600))

	storage := NewStorage(root)

	for _, attempt := range []string{
		"../secret.txt",
		"../../secret.txt",
		"42/../../secret.txt",
	} {
		file, err := storage.Open(attempt)
		if err == nil {
			file.Close()
			t.Fatalf("%q was allowed to escape the storage root", attempt)
		}
	}
}

// Without a directory the application must say so plainly rather than fail partway
// through an upload -- and silently losing photos at the next deploy is the failure
// this guards against.
func TestSave_WithoutADirectoryIsAClearError(t *testing.T) {
	storage := NewStorage("")

	assert.False(t, storage.Configured())

	_, err := storage.Save(1, "front", ".jpg", []byte("x"))
	assert.ErrorIs(t, err, ErrNoStorageDir)
}

func TestExtensionFor_AcceptsOnlyPictures(t *testing.T) {
	for contentType, want := range map[string]string{
		"image/jpeg": ".jpg",
		"image/png":  ".png",
		"image/webp": ".webp",
		"IMAGE/JPEG": ".jpg",
	} {
		extension, ok := ExtensionFor(contentType)
		assert.True(t, ok, contentType)
		assert.Equal(t, want, extension)
	}

	for _, contentType := range []string{"text/html", "application/pdf", "image/svg+xml", ""} {
		_, ok := ExtensionFor(contentType)
		assert.False(t, ok, "%q must not be storable: it is served back to browsers", contentType)
	}
}

// A photo already gone is the state we wanted, so removing it twice is not an error.
func TestRemove_IsIdempotent(t *testing.T) {
	storage := NewStorage(t.TempDir())

	relative, err := storage.Save(7, "back", ".png", []byte("x"))
	assert.NoError(t, err)

	assert.NoError(t, storage.Remove(relative))
	assert.NoError(t, storage.Remove(relative))
}
