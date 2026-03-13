package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/boobam22/gmcl/db"
	"github.com/spf13/cobra"
)

func NewRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <version>",
		Short: "Remove an installed version and unreferenced files",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			version := args[0]
			conn, err := db.Open()
			if err != nil {
				return err
			}
			defer conn.Close()

			tx, err := conn.Begin()
			if err != nil {
				return err
			}
			defer tx.Rollback()
			if _, err := tx.Exec("DELETE FROM t_version_files WHERE version_id = ?", version); err != nil {
				return err
			}
			if _, err := tx.Exec("UPDATE t_versions SET is_installed = FALSE WHERE id = ?", version); err != nil {
				return err
			}
			if _, err := tx.Exec(`UPDATE t_files SET ref_count = (
SELECT COUNT(*) FROM t_version_files vf WHERE vf.path = t_files.path
)`); err != nil {
				return err
			}
			rows, err := tx.Query("SELECT path FROM t_files WHERE ref_count = 0")
			if err != nil {
				return err
			}
			var paths []string
			for rows.Next() {
				var p string
				if err := rows.Scan(&p); err != nil {
					rows.Close()
					return err
				}
				paths = append(paths, p)
			}
			rows.Close()
			if _, err := tx.Exec("DELETE FROM t_files WHERE ref_count = 0"); err != nil {
				return err
			}
			if err := tx.Commit(); err != nil {
				return err
			}
			for _, p := range paths {
				_ = os.Remove(filepath.Join(db.DataDir(), p))
			}
			_ = os.RemoveAll(filepath.Join(db.DataDir(), "versions", version))
			fmt.Fprintf(cmd.OutOrStdout(), "removed %s\n", version)
			return nil
		},
	}
}
