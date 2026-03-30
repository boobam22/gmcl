package cmd

import (
	"strings"

	"github.com/boobam22/gmcl/cli"
	"github.com/spf13/cobra"
)

func NewRemoveCmd(g *cli.Gmcl) *cobra.Command {
	return &cobra.Command{
		Use:   "remove <version>",
		Short: "Remove an installed version and unreferenced files",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			version := args[0]
			return remove(g, version)
		},
	}
}

func remove(g *cli.Gmcl, version string) error {
	if err := validateVersion(g.DB, version, true); err != nil {
		return err
	}

	tx, err := g.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("UPDATE t_versions SET is_installed = 0 WHERE tag = ?", version); err != nil {
		return err
	}

	rows, err := tx.Query(strings.TrimSpace(`
		DELETE FROM t_files
		WHERE path IN (
			SELECT path FROM t_version_files vf1
			WHERE tag = ?
			AND NOT EXISTS (
				SELECT 1 FROM t_version_files vf2
				WHERE vf2.tag != ? AND vf2.path = vf1.path
			)
		)
		RETURNING path
	`), version, version)
	if err != nil {
		return err
	}
	defer rows.Close()

	var orphanFiles []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return err
		}
		orphanFiles = append(orphanFiles, p)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	res, err := tx.Exec("DELETE FROM t_version_files WHERE tag = ?", version)
	if err != nil {
		return err
	}

	linksRemoved, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	for _, p := range orphanFiles {
		g.Remove(p)
	}

	g.Logger.Infof("%d/%d files removed", len(orphanFiles), linksRemoved)
	return nil
}
