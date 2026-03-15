package cmd

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/boobam22/gmcl/db"
	"github.com/spf13/cobra"
)

func NewStartCmd() *cobra.Command {
	var vanilla bool
	cmd := &cobra.Command{
		Use:   "start [version]",
		Short: "Launch Minecraft using local Java runtime",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			conn, err := db.Open()
			if err != nil {
				return err
			}
			defer conn.Close()
			version, err := resolveStartVersion(conn, args)
			if err != nil {
				return err
			}
			if !vanilla {
				fmt.Fprintln(cmd.ErrOrStderr(), "fabric mode requested, falling back to vanilla start")
			}
			if err := runVersion(conn, version); err != nil {
				return err
			}
			_, _ = conn.Exec("UPDATE t_versions SET last_started_at = ? WHERE id = ?", time.Now().UTC().Format(time.RFC3339), version)
			return nil
		},
	}
	cmd.Flags().BoolVar(&vanilla, "vanilla", false, "Disable fabric mode")
	return cmd
}

func resolveStartVersion(conn *sql.DB, args []string) (string, error) {
	if len(args) == 1 {
		return args[0], nil
	}
	var v string
	err := conn.QueryRow("SELECT id FROM t_versions WHERE is_installed = TRUE AND last_started_at IS NOT NULL ORDER BY last_started_at DESC LIMIT 1").Scan(&v)
	if err == nil {
		return v, nil
	}
	err = conn.QueryRow("SELECT id FROM t_versions WHERE is_installed = TRUE ORDER BY release_time DESC LIMIT 1").Scan(&v)
	if err == nil {
		return v, nil
	}
	return "", fmt.Errorf("no installed version found")
}

func runVersion(conn *sql.DB, version string) error {
	metaPath := filepath.Join(db.DataDir(), "versions", version, "metadata.json")
	metaBytes, err := os.ReadFile(metaPath)
	if err != nil {
		return err
	}
	var vm versionMeta
	if err := json.Unmarshal(metaBytes, &vm); err != nil {
		return err
	}
	rows, err := conn.Query("SELECT path FROM t_version_files WHERE version_id = ?", version)
	if err != nil {
		return err
	}
	defer rows.Close()
	var libs []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return err
		}
		if strings.HasPrefix(p, "libraries/") {
			libs = append(libs, filepath.Join(db.DataDir(), p))
		}
	}
	sort.Strings(libs)
	client := filepath.Join(db.DataDir(), "versions", version, "client.jar")
	cp := strings.Join(append(libs, client), string(os.PathListSeparator))
	args := []string{"-cp", cp, vm.MainClass}
	c := exec.Command("java", args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin
	return c.Run()
}
