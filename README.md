# gmcl

gmcl is a lightweight Minecraft launcher written in Go.

It focuses on:

- managing multiple Minecraft versions
- downloading official game assets and libraries
- deduplicating files shared across versions
- generating and executing the correct Java launch command

gmcl uses SQLite to track file references so that multiple versions can share files safely without duplication.

The project is designed to be simple, deterministic, and fully controllable from the command line.

---

# Working Directory

gmcl stores all data inside the `.gmcl` directory.

Example structure:

.gmcl/
  db.sqlite
  versions/
    1.21.5/
      client.jar
      metadata.json
      fabric-metadata.json
      launch.json
  libraries/
  assets/
    indexes/
    objects/

Explanation:

- **db.sqlite**  
  SQLite database tracking installed versions and file references.

- **versions/**  
  Per-version directories containing metadata and files required to launch.

- **libraries/**  
  Shared libraries downloaded from Mojang.

- **assets/**  
  Minecraft asset storage following the official structure.

---

# Data Sources

gmcl uses Mojang’s official metadata API.

Main version manifest:

https://piston-meta.mojang.com/mc/game/version_manifest.json

Each version entry references another JSON document that describes:

- downloads
- libraries
- asset index
- launch arguments

---

# Commands

## update

Downloads the latest version manifest and updates the local database.

This command **does not install any game files**.

---

## list

Lists versions known to the local database.

Options:

- `--type release`
- `--type snapshot`

Pattern is an optional glob filter.

Examples:

```sh
gmcl list
gmcl list --type release
gmcl list "1.20*"
```

---

## install

Installs a specific Minecraft version.

Steps:

1. Verify the version exists in the database.
2. Download the version metadata.
3. Parse asset and library dependencies.
4. Record file references in SQLite.
5. Download missing files.

Downloads are performed using **32 concurrent workers**.

Files are verified using SHA1 hashes from the official metadata.

Assets are stored using their SHA1-based object layout.
Libraries follow Mojang's Maven-style directory layout.

---

## remove

Removes an installed version.

Behavior:

1. Remove the version record from the database.
2. Check file reference counts.
3. Delete files that are no longer referenced.

Shared files used by other versions are preserved.

---

## start

Launches Minecraft using the system Java runtime.

Version resolution:

1. Use explicitly specified version
2. Otherwise use the last started version
3. Otherwise use the latest installed version

gmcl generates the correct Java launch command and executes it.

---

# Fabric Mode

Fabric is enabled by default.

To disable Fabric:

```sh
gmcl start --vanilla
```

When Fabric mode is active:

- Fabric loader libraries are included
- Fabric entry class is used

Example entry class:

```sh
net.fabricmc.loader.impl.launch.knot.KnotClient
```

---

# Java Runtime

gmcl does not manage Java installations.

It assumes that `java` is available in the system PATH.

Users are responsible for installing and managing their Java runtime.
