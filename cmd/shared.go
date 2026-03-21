package cmd

import (
	"database/sql"
	"errors"
	"fmt"
)

const timeFormat = "2006-01-02 15:04:05"

type versionManifest struct {
	Latest struct {
		Release  string `json:"release"`
		Snapshot string `json:"snapshot"`
	} `json:"latest"`
	Versions []struct {
		Tag         string `json:"id"`
		Type        string `json:"type"`
		URL         string `json:"url"`
		Time        string `json:"time"`
		ReleaseTime string `json:"releaseTime"`
	} `json:"versions"`
}

type versionMeta struct {
	Downloads struct {
		Client struct {
			URL  string `json:"url"`
			Size int64  `json:"size"`
			SHA1 string `json:"sha1"`
		} `json:"client"`
	} `json:"downloads"`
	Libraries []struct {
		Downloads struct {
			Artifact struct {
				Path string `json:"path"`
				URL  string `json:"url"`
				Size int64  `json:"size"`
				SHA1 string `json:"sha1"`
			} `json:"artifact"`
		} `json:"downloads"`
		Name string `json:"name"`
	} `json:"libraries"`
	AssetIndex struct {
		ID   string `json:"id"`
		URL  string `json:"url"`
		Size int64  `json:"size"`
		SHA1 string `json:"sha1"`
	} `json:"assetIndex"`
}

type assetIndex struct {
	Objects map[string]struct {
		Size int64  `json:"size"`
		SHA1 string `json:"hash"`
	} `json:"objects"`
}

type fabricVersions []struct {
	Loader struct {
		Version string `json:"version"`
	} `json:"loader"`
}

type fabricMeta struct {
	Libraries []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
		Size int64  `json:"size"`
		SHA1 string `json:"sha1"`
	} `json:"libraries"`
}

func validateVersion(db *sql.DB, version string, expected bool) error {
	var installed bool
	if err := db.QueryRow("SELECT is_installed FROM t_versions WHERE tag = ?", version).Scan(&installed); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("version %s not found", version)
		}
		return err
	}

	if installed == expected {
		return nil
	}

	if installed {
		return fmt.Errorf("version %s installed", version)
	} else {
		return fmt.Errorf("version %s not installed", version)
	}
}
