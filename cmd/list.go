package cmd

import (
	"database/sql"
	"fmt"
	"path"

	"github.com/boobam22/gmcl/db"
	"github.com/spf13/cobra"
)

func NewListCmd() *cobra.Command {
	var t string
	cmd := &cobra.Command{
		Use:   "list [pattern]",
		Short: "List versions available in local database",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			conn, err := db.Open()
			if err != nil {
				return err
			}
			defer conn.Close()

			query := "SELECT id, type, is_installed FROM t_versions"
			var rows *sql.Rows
			if t != "" {
				query += " WHERE type = ?"
			}
			query += " ORDER BY release_time DESC"
			if t != "" {
				rows, err = conn.Query(query, t)
			} else {
				rows, err = conn.Query(query)
			}
			if err != nil {
				return err
			}
			defer rows.Close()

			pattern := "*"
			if len(args) == 1 {
				pattern = args[0]
			}
			for rows.Next() {
				var id, typ string
				var installed bool
				if err := rows.Scan(&id, &typ, &installed); err != nil {
					return err
				}
				if ok, _ := path.Match(pattern, id); !ok {
					continue
				}
				installedMark := " "
				if installed {
					installedMark = "*"
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%s %s (%s)\n", installedMark, id, typ)
			}
			return rows.Err()
		},
	}
	cmd.Flags().StringVar(&t, "type", "", "Filter by version type (release/snapshot)")
	return cmd
}
