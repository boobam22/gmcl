package cmd

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/boobam22/gmcl/db"
)

const versionManifestURL = "https://piston-meta.mojang.com/mc/game/version_manifest.json"

func ensureDirs() error {
	dirs := []string{
		db.DataDir(),
		filepath.Join(db.DataDir(), "versions"),
		filepath.Join(db.DataDir(), "libraries"),
		filepath.Join(db.DataDir(), "assets", "indexes"),
		filepath.Join(db.DataDir(), "assets", "objects"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return err
		}
	}
	return nil
}

func downloadFile(url, path string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("download %s: status %s", url, resp.Status)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	return err
}

func sha1File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha1.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func downloadIfNeeded(url, path, expectedSHA string) error {
	if st, err := os.Stat(path); err == nil && st.Mode().IsRegular() {
		if expectedSHA == "" {
			return nil
		}
		hash, err := sha1File(path)
		if err == nil && hash == expectedSHA {
			return nil
		}
	}
	return downloadFile(url, path)
}
