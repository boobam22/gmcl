package cli

import (
	"fmt"
	"net/http"
)

func (g *Gmcl) DownloadFile(url, path string) error {
	g.Logger.Debugf("Downloading %s", path)

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("download %s: %w", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download %s: %s", path, resp.Status)
	}

	written, err := g.WriteToFile(path, resp.Body)
	if err != nil {
		return fmt.Errorf("download %s: %w", path, err)
	}

	g.Logger.Debugf("Downloaded %s (%d bytes)", path, written)
	return nil
}

func (g *Gmcl) DownloadFileIfNeed(url, path string) error {
	if ok, _ := g.Exists(path); ok {
		g.Logger.Debugf("File exists: %s", path)
		return nil
	}

	return g.DownloadFile(url, path)
}
