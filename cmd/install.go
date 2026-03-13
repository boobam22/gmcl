package cmd

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/boobam22/gmcl/db"
	"github.com/spf13/cobra"
)

type versionMeta struct {
	ID        string `json:"id"`
	MainClass string `json:"mainClass"`
	Downloads struct {
		Client struct {
			SHA1 string `json:"sha1"`
			URL  string `json:"url"`
			Size int64  `json:"size"`
		} `json:"client"`
	} `json:"downloads"`
	Libraries []struct {
		Downloads struct {
			Artifact *struct {
				Path string `json:"path"`
				SHA1 string `json:"sha1"`
				URL  string `json:"url"`
				Size int64  `json:"size"`
			} `json:"artifact"`
			Classifiers map[string]struct {
				Path string `json:"path"`
				SHA1 string `json:"sha1"`
				URL  string `json:"url"`
				Size int64  `json:"size"`
			} `json:"classifiers"`
		} `json:"downloads"`
	} `json:"libraries"`
	AssetIndex struct {
		ID   string `json:"id"`
		SHA1 string `json:"sha1"`
		URL  string `json:"url"`
		Size int64  `json:"size"`
	} `json:"assetIndex"`
}

type assetIndex struct {
	Objects map[string]struct {
		Hash string `json:"hash"`
		Size int64  `json:"size"`
	} `json:"objects"`
}

type dlItem struct {
	url, path, sha, kind string
	size                 int64
}

func NewInstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "install <version>",
		Short: "Install a specific Minecraft version",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			version := args[0]
			if err := ensureDirs(); err != nil {
				return err
			}
			conn, err := db.Open()
			if err != nil {
				return err
			}
			defer conn.Close()

			var url string
			if err := conn.QueryRow("SELECT url FROM t_versions WHERE id = ?", version).Scan(&url); err != nil {
				if err == sql.ErrNoRows {
					return fmt.Errorf("version %q not found in database; run gmcl update first", version)
				}
				return err
			}

			metaPath := filepath.Join(db.DataDir(), "versions", version, "metadata.json")
			if err := downloadIfNeeded(url, metaPath, ""); err != nil {
				return err
			}
			metaBytes, err := os.ReadFile(metaPath)
			if err != nil {
				return err
			}
			var vm versionMeta
			if err := json.Unmarshal(metaBytes, &vm); err != nil {
				return err
			}

			items := make([]dlItem, 0, 1024)
			clientPath := filepath.Join("versions", version, "client.jar")
			items = append(items, dlItem{url: vm.Downloads.Client.URL, path: clientPath, sha: vm.Downloads.Client.SHA1, size: vm.Downloads.Client.Size, kind: "client"})

			for _, lib := range vm.Libraries {
				if lib.Downloads.Artifact != nil {
					a := lib.Downloads.Artifact
					items = append(items, dlItem{url: a.URL, path: filepath.Join("libraries", a.Path), sha: a.SHA1, size: a.Size, kind: "library"})
				}
				for classifier, c := range lib.Downloads.Classifiers {
					if strings.Contains(classifier, runtime.GOOS) {
						items = append(items, dlItem{url: c.URL, path: filepath.Join("libraries", c.Path), sha: c.SHA1, size: c.Size, kind: "native"})
					}
				}
			}

			idxPath := filepath.Join(db.DataDir(), "assets", "indexes", vm.AssetIndex.ID+".json")
			if err := downloadIfNeeded(vm.AssetIndex.URL, idxPath, vm.AssetIndex.SHA1); err != nil {
				return err
			}
			items = append(items, dlItem{url: vm.AssetIndex.URL, path: filepath.Join("assets", "indexes", vm.AssetIndex.ID+".json"), sha: vm.AssetIndex.SHA1, size: vm.AssetIndex.Size, kind: "asset-index"})
			var ai assetIndex
			idxBytes, err := os.ReadFile(idxPath)
			if err != nil {
				return err
			}
			if err := json.Unmarshal(idxBytes, &ai); err != nil {
				return err
			}
			for _, obj := range ai.Objects {
				sub := obj.Hash[:2]
				p := filepath.Join("assets", "objects", sub, obj.Hash)
				u := fmt.Sprintf("https://resources.download.minecraft.net/%s/%s", sub, obj.Hash)
				items = append(items, dlItem{url: u, path: p, sha: obj.Hash, size: obj.Size, kind: "asset"})
			}

			if err := downloadItems(items); err != nil {
				return err
			}
			if err := writeRefs(conn, version, items); err != nil {
				return err
			}
			_, err = conn.Exec("UPDATE t_versions SET is_installed = TRUE WHERE id = ?", version)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "installed %s (%d files)\n", version, len(items))
			return nil
		},
	}
}

func downloadItems(items []dlItem) error {
	ch := make(chan dlItem)
	errCh := make(chan error, 1)
	var wg sync.WaitGroup
	worker := func() {
		defer wg.Done()
		for it := range ch {
			fullPath := filepath.Join(db.DataDir(), it.path)
			if err := downloadIfNeeded(it.url, fullPath, it.sha); err != nil {
				select {
				case errCh <- err:
				default:
				}
				return
			}
		}
	}
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go worker()
	}
	for _, it := range items {
		select {
		case err := <-errCh:
			close(ch)
			wg.Wait()
			return err
		default:
			ch <- it
		}
	}
	close(ch)
	wg.Wait()
	select {
	case err := <-errCh:
		return err
	default:
		return nil
	}
}

func writeRefs(conn *sql.DB, version string, items []dlItem) error {
	tx, err := conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec("DELETE FROM t_version_files WHERE version_id = ?", version); err != nil {
		return err
	}
	for _, it := range items {
		if _, err := tx.Exec(`INSERT INTO t_files(path, sha, size, kind, ref_count) VALUES(?, ?, ?, ?, 0)
ON CONFLICT(path) DO UPDATE SET sha=excluded.sha, size=excluded.size, kind=excluded.kind`, it.path, it.sha, it.size, it.kind); err != nil {
			return err
		}
		if _, err := tx.Exec("INSERT OR IGNORE INTO t_version_files(version_id, path) VALUES(?, ?)", version, it.path); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(`UPDATE t_files SET ref_count = (
SELECT COUNT(*) FROM t_version_files vf WHERE vf.path = t_files.path
)`); err != nil {
		return err
	}
	return tx.Commit()
}
