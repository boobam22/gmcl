package cli

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
)

func (g *Gmcl) LoadJSON(p string, v any) error {
	bytes, err := os.ReadFile(filepath.Join(g.DataDir, p))
	if err != nil {
		return err
	}

	if err := json.Unmarshal(bytes, v); err != nil {
		return err
	}

	return nil
}

func (g *Gmcl) Exists(p string) (bool, error) {
	if _, err := os.Stat(filepath.Join(g.DataDir, p)); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (g *Gmcl) Remove(p string) error {
	err := os.Remove(filepath.Join(g.DataDir, p))
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (g *Gmcl) WriteToFile(p string, src io.Reader) (int64, error) {
	fullPath := filepath.Join(g.DataDir, p)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return 0, err
	}

	tmpPath := fullPath + ".tmp"
	defer os.Remove(tmpPath)

	dst, err := os.Create(tmpPath)
	if err != nil {
		return 0, err
	}
	defer dst.Close()

	written, err := io.Copy(dst, src)
	if err != nil {
		return 0, err
	}

	if err = os.Rename(tmpPath, fullPath); err != nil {
		return 0, err
	}

	return written, nil
}
