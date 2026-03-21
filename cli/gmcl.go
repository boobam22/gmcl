package cli

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath"
)

type Gmcl struct {
	DB     *sql.DB
	Logger *Logger

	DataDir     string
	AssetDir    string
	AssetIdxDir string
	AssetObjDir string
	VersionDir  string
	LibraryDir  string
	NativeDir   string
}

func NewGmcl() (*Gmcl, error) {
	home, _ := os.UserHomeDir()
	dataDir := filepath.Join(home, ".gmcl")

	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return nil, err
	}

	dbConn, err := NewDB(filepath.Join(dataDir, "db.sqlite"))
	if err != nil {
		return nil, err
	}

	logger, err := NewLogger(filepath.Join(dataDir, "gmcl.log"))
	if err != nil {
		return nil, err
	}

	return &Gmcl{
		DB:     dbConn,
		Logger: logger,

		DataDir:     dataDir,
		AssetDir:    "assets",
		AssetIdxDir: "assets/indexes",
		AssetObjDir: "assets/objects",
		VersionDir:  "versions",
		LibraryDir:  "libraries",
		NativeDir:   "natives",
	}, nil
}

func (g *Gmcl) Close() error {
	var err1, err2 error

	if g.DB != nil {
		err1 = g.DB.Close()
	}

	if g.Logger != nil {
		err2 = g.Logger.Close()
	}

	return errors.Join(err1, err2)
}
