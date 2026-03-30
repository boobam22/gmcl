package cmd

import (
	"fmt"
	"io"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/boobam22/gmcl/cli"
	"github.com/spf13/cobra"
)

func NewStartCmd(g *cli.Gmcl) *cobra.Command {
	var vanilla bool

	cmd := &cobra.Command{
		Use:   "start [version]",
		Short: "Launch Minecraft using local Java runtime",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			version := ""
			if len(args) == 1 {
				version = args[0]
			}

			return start(g, version, vanilla)
		},
	}

	cmd.Flags().BoolVar(&vanilla, "vanilla", false, "Disable fabric mode")

	return cmd
}

func start(g *cli.Gmcl, version string, vanilla bool) error {
	if version == "" {
		var err error
		version, err = selectFallbackVersion(g)
		if err != nil {
			return err
		}
	} else {
		if err := validateVersion(g.DB, version, true); err != nil {
			return err
		}
	}

	_, err := g.DB.Exec(
		"UPDATE t_versions SET last_started_at = ? WHERE tag = ?",
		time.Now().Local().Format(timeFormat),
		version,
	)
	if err != nil {
		return err
	}

	vanillaArgFile, err := ensureVanillaArgFile(g, version)
	if err != nil {
		return err
	}
	fabricArgFile, err := ensureFabricArgFile(g, version)
	if err != nil {
		return err
	}

	argFile := fabricArgFile
	if vanilla {
		argFile = vanillaArgFile
	}

	cmd := exec.Command("java", "@"+argFile)
	cmd.Dir = g.DataDir
	cmd.Stdout = g.Logger.File
	cmd.Stderr = g.Logger.File

	return cmd.Start()
}

func selectFallbackVersion(g *cli.Gmcl) (string, error) {
	query := strings.TrimSpace(`
		SELECT tag FROM t_versions
		WHERE is_installed = 1
		ORDER BY
			CASE WHEN last_started_at IS NOT NULL THEN 0 ELSE 1 END,
			COALESCE(last_started_at, release_time) DESC
		LIMIT 1
	`)

	var version string
	if err := g.DB.QueryRow(query).Scan(&version); err != nil {
		return "", fmt.Errorf("no installed version found to start")
	}

	return version, nil
}

func ensureVanillaArgFile(g *cli.Gmcl, version string) (string, error) {
	p := path.Join(g.VersionDir, version, "vanilla.arg")
	if ok, _ := g.Exists(p); ok {
		return p, nil
	}

	var sb strings.Builder
	writeJVMArg(g, &sb, version, true)
	writeGameArg(g, &sb, version)

	g.WriteToFile(p, strings.NewReader(sb.String()))
	return p, nil
}

func ensureFabricArgFile(g *cli.Gmcl, version string) (string, error) {
	p := path.Join(g.VersionDir, version, "fabric.arg")
	if ok, _ := g.Exists(p); ok {
		return p, nil
	}

	var sb strings.Builder
	writeJVMArg(g, &sb, version, false)
	writeGameArg(g, &sb, version)

	if _, err := g.WriteToFile(p, strings.NewReader(sb.String())); err != nil {
		return "", err
	}

	return p, nil
}

func writeJVMArg(g *cli.Gmcl, w io.Writer, version string, vanilla bool) {
	if vanilla {
		fmt.Fprintln(w, "-Xms2G -Xmx4G")
	} else {
		fmt.Fprintln(w, "-Xms3G -Xmx8G")
	}

	nativeDir := path.Join(g.VersionDir, version, g.NativeDir)
	fmt.Fprintf(w, "-Djava.library.path=%s\n", nativeDir)
	fmt.Fprintf(w, "-Djna.tmpdir=%s\n", nativeDir)
	fmt.Fprintf(w, "-Dorg.lwjgl.system.SharedLibraryExtractPath=%s\n", nativeDir)
	fmt.Fprintf(w, "-Dio.netty.native.workdir=%s\n", nativeDir)
	fmt.Fprintf(w, "-cp %s\n", path.Join(g.VersionDir, version, g.LibraryDir, "*"))

	if vanilla {
		fmt.Fprintln(w, "net.minecraft.client.main.Main")
	} else {
		fmt.Fprintln(w, "net.fabricmc.loader.impl.launch.knot.KnotClient")
	}
}

func writeGameArg(g *cli.Gmcl, w io.Writer, version string) error {
	var vm versionMeta
	if err := g.LoadJSON(path.Join(g.VersionDir, version, "metadata.json"), &vm); err != nil {
		return err
	}

	fmt.Fprintf(w, "--username %s\n", "nia11720")
	fmt.Fprintf(w, "--version %s\n", version)
	fmt.Fprintf(w, "--gameDir %s\n", path.Join(g.VersionDir, version))
	fmt.Fprintf(w, "--assetsDir %s\n", path.Join(g.AssetDir))
	fmt.Fprintf(w, "--assetIndex %s\n", vm.AssetIndex.ID)
	fmt.Fprintf(w, "--uuid %s\n", "3b6de038-3d66-48cd-8816-cb3579dd2c53")
	fmt.Fprintf(w, "--accessToken %s\n", "0")

	return nil
}
