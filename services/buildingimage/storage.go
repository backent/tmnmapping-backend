// Package buildingimage stores building photos this application hosts, and decides
// when to fall back to the ones ERP serves.
package buildingimage

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Storage writes photo files under a single root directory.
//
// That directory is a VOLUME on the host, not part of the image: the container
// filesystem is replaced on every deploy, and this one is deployed many times a day.
// Point BUILDING_IMAGES_DIR somewhere that survives, or uploads vanish at the next
// release without anything reporting an error.
type Storage struct {
	root string
}

// ErrNoStorageDir means the application was started without somewhere to put photos.
var ErrNoStorageDir = errors.New("BUILDING_IMAGES_DIR is not set, so uploaded photos have nowhere to live")

// MaxUploadBytes caps one photo. Generous for a building exterior, small enough that
// a mistaken upload cannot fill the volume.
const MaxUploadBytes = 8 << 20 // 8 MiB

// allowedContentTypes is a whitelist, not a blacklist: the file is served back to
// browsers, so anything not obviously a picture stays out.
var allowedContentTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

func NewStorage(root string) *Storage {
	return &Storage{root: strings.TrimSpace(root)}
}

// NewStorageFromEnv reads BUILDING_IMAGES_DIR.
func NewStorageFromEnv() *Storage {
	return NewStorage(os.Getenv("BUILDING_IMAGES_DIR"))
}

// Configured reports whether there is anywhere to write. Checked before an upload is
// accepted, so the caller gets a clear message rather than a write error halfway.
func (s *Storage) Configured() bool { return s.root != "" }

// ExtensionFor validates the content type and returns the extension to store under.
func ExtensionFor(contentType string) (string, bool) {
	extension, ok := allowedContentTypes[strings.ToLower(strings.TrimSpace(contentType))]

	return extension, ok
}

// AllowedContentTypes lists what an upload may be, for the error message and the
// file picker's accept attribute.
func AllowedContentTypes() []string {
	return []string{"image/jpeg", "image/png", "image/webp"}
}

// Save writes one photo and returns its path relative to the root.
//
// The name carries random bytes rather than being <building>/<slot>.<ext>: a browser
// that cached the old photo would otherwise keep showing it after a replacement,
// since the URL would not have changed.
func (s *Storage) Save(buildingId int, slot string, extension string, content []byte) (string, error) {
	if !s.Configured() {
		return "", ErrNoStorageDir
	}

	suffix := make([]byte, 8)
	if _, err := rand.Read(suffix); err != nil {
		return "", err
	}

	relative := filepath.Join(
		strconv.Itoa(buildingId),
		fmt.Sprintf("%s-%s%s", slot, hex.EncodeToString(suffix), extension),
	)

	full := filepath.Join(s.root, relative)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return "", err
	}

	if err := os.WriteFile(full, content, 0o644); err != nil {
		return "", err
	}

	return relative, nil
}

// Open returns the stored file for serving.
func (s *Storage) Open(relative string) (*os.File, error) {
	full, err := s.resolve(relative)
	if err != nil {
		return nil, err
	}

	return os.Open(full)
}

// Remove deletes a stored file. A missing file is not an error: the row is going away
// either way, and a photo already gone is the state we wanted.
func (s *Storage) Remove(relative string) error {
	full, err := s.resolve(relative)
	if err != nil {
		return err
	}

	if err := os.Remove(full); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

// resolve turns a stored relative path into a real one, refusing anything that climbs
// out of the root. The paths come from our own database, but a traversal bug here
// would serve arbitrary files off the volume, so it is checked rather than assumed.
func (s *Storage) resolve(relative string) (string, error) {
	if !s.Configured() {
		return "", ErrNoStorageDir
	}

	full := filepath.Join(s.root, filepath.Clean("/"+relative))
	if !strings.HasPrefix(full, filepath.Clean(s.root)+string(os.PathSeparator)) {
		return "", errors.New("image path escapes the storage directory")
	}

	return full, nil
}
