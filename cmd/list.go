package cmd

import (
	"fmt"
	"path"
	"strings"

	"github.com/boobam22/gmcl/cli"
	"github.com/spf13/cobra"
)

type listOptions struct {
	Pattern   string
	Release   bool
	Snapshot  bool
	Installed bool
}

func NewListCmd(g *cli.Gmcl) *cobra.Command {
	var opts listOptions

	cmd := &cobra.Command{
		Use:   "list [pattern]",
		Short: "List versions available in local database",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Pattern = "*"
			if len(args) == 1 {
				opts.Pattern = args[0]
			}

			return list(g, opts)
		},
	}

	cmd.Flags().BoolVar(&opts.Release, "release", false, "Show only release versions")
	cmd.Flags().BoolVar(&opts.Snapshot, "snapshot", false, "Show only snapshot versions")
	cmd.Flags().BoolVar(&opts.Installed, "installed", false, "Show only installed versions")

	cmd.MarkFlagsMutuallyExclusive("release", "snapshot")

	return cmd
}

func list(g *cli.Gmcl, opts listOptions) error {
	query, args := buildQuery(opts)

	rows, err := g.DB.Query(query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	formt := "%1s %-24s %-12s %s\n"
	fmt.Printf(formt, "I", "TAG", "TYPE", "RELEASE TIME")

	for rows.Next() {
		var tag, versionType, releaseTime string
		var installed bool

		if err := rows.Scan(&tag, &versionType, &releaseTime, &installed); err != nil {
			return err
		}

		ok, err := path.Match(opts.Pattern, tag)
		if err != nil {
			return err
		}
		if !ok {
			continue
		}

		mark := ""
		if installed {
			mark = "*"
		}

		fmt.Printf(formt, mark, tag, versionType, releaseTime)
	}

	return rows.Err()
}

func buildQuery(opts listOptions) (string, []any) {
	query := "SELECT tag, type, release_time, is_installed FROM t_versions"

	conds := make([]string, 0, 3)
	args := make([]any, 0, 3)

	if opts.Release {
		conds = append(conds, "type = ?")
		args = append(args, "release")
	} else if opts.Snapshot {
		conds = append(conds, "type = ?")
		args = append(args, "snapshot")
	}

	if opts.Installed {
		conds = append(conds, "is_installed = ?")
		args = append(args, 1)
	}

	if len(conds) > 0 {
		query += " WHERE " + strings.Join(conds, " AND ")
	}

	query += " ORDER BY release_time DESC"

	return query, args
}
