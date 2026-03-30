package cmd

import (
	"fmt"
	"net/url"
	"path"
	"strings"
	"sync"

	"github.com/boobam22/gmcl/cli"
	"github.com/spf13/cobra"
)

func NewInstallCmd(g *cli.Gmcl) *cobra.Command {
	return &cobra.Command{
		Use:   "install <version>",
		Short: "Install a specific Minecraft version",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return install(g, args[0])
		},
	}
}

func install(g *cli.Gmcl, version string) error {
	if err := validateVersion(g.DB, version, false); err != nil {
		return err
	}

	g.Logger.Info("Resolving core files...")

	items, err := resolveCoreFiles(g, version)
	if err != nil {
		return err
	}

	g.Logger.Info("Resolving fabric...")

	fabricLibs, err := resolveFabric(g, version)
	if err != nil {
		return err
	}

	items = append(items, fabricLibs...)

	g.Logger.Infof("Downloading libraries and assets (%d files)...", len(items))

	if err := downloadMany(g, items); err != nil {
		return err
	}

	if err := writeToDB(g, version, items); err != nil {
		return err
	}

	g.Logger.Infof("Version %s is installed", version)
	return nil
}

type dlItem struct {
	URL, Path, Name, SHA1 string
	Size                  int64
}

func resolveCoreFiles(g *cli.Gmcl, version string) ([]dlItem, error) {
	var u string
	if err := g.DB.QueryRow("SELECT url FROM t_versions WHERE tag = ?", version).Scan(&u); err != nil {
		return nil, err
	}

	var vm versionMeta
	metaPath := path.Join(g.VersionDir, version, "metadata.json")
	if err := g.DownloadFileIfNeed(u, metaPath); err != nil {
		return nil, err
	}
	if err := g.LoadJSON(metaPath, &vm); err != nil {
		return nil, err
	}

	var ai assetIndex
	idxPath := path.Join(g.AssetIdxDir, vm.AssetIndex.ID+".json")
	if err := g.DownloadFileIfNeed(vm.AssetIndex.URL, idxPath); err != nil {
		return nil, err
	}
	if err := g.LoadJSON(idxPath, &ai); err != nil {
		return nil, err
	}

	clientPath := path.Join(g.VersionDir, version, g.LibraryDir, "client.jar")
	if err := g.DownloadFileIfNeed(vm.Downloads.Client.URL, clientPath); err != nil {
		return nil, err
	}

	items := make([]dlItem, 0, 5000)

	for _, lib := range vm.Libraries {
		a := lib.Downloads.Artifact
		items = append(items, dlItem{
			URL:  a.URL,
			Path: path.Join(g.VersionDir, version, g.LibraryDir, flattenMavenPath(lib.Name)),
			Size: a.Size,
			Name: lib.Name,
			SHA1: a.SHA1,
		})
	}

	for name, obj := range ai.Objects {
		sha1 := obj.SHA1
		sub := sha1[:2]

		u, err := url.JoinPath("https://resources.download.minecraft.net", sub, sha1)
		if err != nil {
			return nil, err
		}

		items = append(items, dlItem{
			URL:  u,
			Path: path.Join(g.AssetObjDir, sub, sha1),
			Size: obj.Size,
			Name: name,
			SHA1: sha1,
		})
	}

	return items, nil
}

func resolveFabric(g *cli.Gmcl, gameVersion string) ([]dlItem, error) {
	fabricHost := "https://meta.fabricmc.net"
	versionsURL := fmt.Sprintf("%s/v2/versions/loader/%s", fabricHost, gameVersion)
	tmp := "fabric-loader.json"

	if err := g.DownloadFile(versionsURL, tmp); err != nil {
		return nil, err
	}
	defer g.Remove(tmp)

	var versions fabricVersions
	if err := g.LoadJSON(tmp, &versions); err != nil {
		return nil, err
	}

	if len(versions) == 0 {
		return nil, fmt.Errorf("no fabric versions found for %s", gameVersion)
	}

	metaURL := fmt.Sprintf("%s/v2/versions/loader/%s/%s/profile/json", fabricHost, gameVersion, versions[0].Loader.Version)
	metaPath := path.Join(g.VersionDir, gameVersion, "fabric-metadata.json")
	if err := g.DownloadFileIfNeed(metaURL, metaPath); err != nil {
		return nil, err
	}

	var fm fabricMeta
	if err := g.LoadJSON(metaPath, &fm); err != nil {
		return nil, err
	}

	items := make([]dlItem, 0, 128)

	for _, it := range fm.Libraries {
		p, err := mavenNameToPath(it.Name)
		if err != nil {
			return nil, err
		}

		u, err := url.JoinPath(it.URL, p)
		if err != nil {
			return nil, err
		}

		items = append(items, dlItem{
			URL:  u,
			Path: path.Join(g.VersionDir, gameVersion, g.LibraryDir, flattenMavenPath(it.Name)),
			Size: it.Size,
			Name: it.Name,
			SHA1: it.SHA1,
		})
	}

	return items, nil
}

func downloadMany(g *cli.Gmcl, items []dlItem) error {
	sem := make(chan struct{}, 16)
	var wg sync.WaitGroup

	failedCount := 0
	var mu sync.Mutex

	for _, item := range items {
		wg.Add(1)
		sem <- struct{}{}

		go func(it dlItem) {
			defer wg.Done()
			defer func() { <-sem }()

			if err := g.DownloadFileIfNeed(it.URL, it.Path); err != nil {
				mu.Lock()
				failedCount++
				mu.Unlock()
				g.Logger.Error(err.Error())
			}
		}(item)
	}

	wg.Wait()

	if failedCount > 0 {
		return fmt.Errorf("%d files download failed", failedCount)
	}
	return nil
}

func writeToDB(g *cli.Gmcl, version string, items []dlItem) error {
	tx, err := g.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmtFile, err := tx.Prepare("INSERT OR IGNORE INTO t_files(path, size, name, sha1) VALUES(?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmtFile.Close()

	stmtVersionFile, err := tx.Prepare("INSERT OR IGNORE INTO t_version_files(tag, path) VALUES(?, ?)")
	if err != nil {
		return err
	}
	defer stmtVersionFile.Close()

	for _, it := range items {
		if strings.HasPrefix(it.Path, g.VersionDir) {
			continue
		}

		if _, err := stmtFile.Exec(it.Path, it.Size, it.Name, it.SHA1); err != nil {
			return err
		}
		if _, err := stmtVersionFile.Exec(version, it.Path); err != nil {
			return err
		}
	}

	if _, err := tx.Exec("UPDATE t_versions SET is_installed = 1, last_started_at = NULL WHERE tag = ?", version); err != nil {
		return err
	}

	return tx.Commit()
}

func mavenNameToPath(n string) (string, error) {
	parts := strings.Split(n, ":")

	if len(parts) != 3 {
		return "", fmt.Errorf("invalid maven name: %q (expected format group:artifact:version)", n)
	}

	group := strings.ReplaceAll(parts[0], ".", "/")
	artifact := parts[1]
	version := parts[2]

	if group == "" || artifact == "" || version == "" {
		return "", fmt.Errorf("invalid maven name: %q (empty component)", n)
	}

	return fmt.Sprintf("%s/%s/%s/%s-%s.jar", group, artifact, version, artifact, version), nil
}

func flattenMavenPath(n string) string {
	return strings.ReplaceAll(n, ":", "-") + ".jar"
}
