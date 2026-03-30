package cmd

import (
	"strings"
	"time"

	"github.com/boobam22/gmcl/cli"
	"github.com/spf13/cobra"
)

func NewUpdateCmd(g *cli.Gmcl) *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Update local version metadata from Mojang",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return update(g)
		},
	}
}

func update(g *cli.Gmcl) error {
	g.Logger.Info("Starting Minecraft version manifest update...")

	url := "https://piston-meta.mojang.com/mc/game/version_manifest.json"
	path := "version_manifest.json"
	if err := g.DownloadFile(url, path); err != nil {
		return err
	}

	var manifest versionManifest
	if err := g.LoadJSON(path, &manifest); err != nil {
		return err
	}

	g.Logger.Infof(
		"%d total versions, latest release=%s, latest snapshot=%s",
		len(manifest.Versions),
		manifest.Latest.Release,
		manifest.Latest.Snapshot,
	)

	tx, err := g.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(strings.TrimSpace(`
		INSERT INTO t_versions(tag, type, url, time, release_time, is_installed)
		VALUES(?, ?, ?, ?, ?, 0)
		ON CONFLICT(tag) DO UPDATE
		SET type=excluded.type,
			url=excluded.url,
			time=excluded.time,
			release_time=excluded.release_time
		WHERE type != excluded.type
			OR url != excluded.url
			OR time != excluded.time
			OR release_time != excluded.release_time
	`))
	if err != nil {
		return err
	}
	defer stmt.Close()

	syncedCount := 0
	for _, v := range manifest.Versions {
		vTime, err := formatTimeLocal(v.Time)
		if err != nil {
			return err
		}
		vReleaseTime, err := formatTimeLocal(v.ReleaseTime)
		if err != nil {
			return err
		}

		res, err := stmt.Exec(v.Tag, v.Type, v.URL, vTime, vReleaseTime)
		if err != nil {
			return err
		}

		affected, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if affected == 1 {
			syncedCount++
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	g.Logger.Infof("Version update complete: %d versions synced successfully.", syncedCount)
	return nil
}

func formatTimeLocal(s string) (string, error) {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return "", err
	}

	return t.Local().Format(timeFormat), nil
}
