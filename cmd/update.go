package cmd

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/boobam22/gmcl/db"
	"github.com/spf13/cobra"
)

type versionManifest struct {
	Latest   map[string]string `json:"latest"`
	Versions []struct {
		ID          string `json:"id"`
		Type        string `json:"type"`
		URL         string `json:"url"`
		Time        string `json:"time"`
		ReleaseTime string `json:"releaseTime"`
	} `json:"versions"`
}

func NewUpdateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Update local version metadata from Mojang",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := ensureDirs(); err != nil {
				return err
			}
			conn, err := db.Open()
			if err != nil {
				return err
			}
			defer conn.Close()
			manifest, err := fetchManifest()
			if err != nil {
				return err
			}
			if err := upsertVersions(conn, manifest); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "updated %d versions\n", len(manifest.Versions))
			return nil
		},
	}
}

func fetchManifest() (*versionManifest, error) {
	resp, err := http.Get(versionManifestURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("fetch manifest: %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var m versionManifest
	if err := json.Unmarshal(body, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

func upsertVersions(conn *sql.DB, manifest *versionManifest) error {
	tx, err := conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, v := range manifest.Versions {
		if _, err := tx.Exec(`
INSERT INTO t_versions(id, type, time, release_time, url, is_installed)
VALUES(?, ?, ?, ?, ?, COALESCE((SELECT is_installed FROM t_versions WHERE id=?), FALSE))
ON CONFLICT(id) DO UPDATE SET type=excluded.type, time=excluded.time, release_time=excluded.release_time, url=excluded.url`,
			v.ID, v.Type, v.Time, v.ReleaseTime, v.URL, v.ID); err != nil {
			return err
		}
	}
	return tx.Commit()
}
