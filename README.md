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

## Working Directory

gmcl stores all data inside the `.gmcl` directory.

Example structure:

.gmcl/
  db.sqlite
  gmcl.log
  version_manifest.json
  versions/
    1.21.5/
      libraries/
        client.jar
      natives/
      metadata.json
      fabric-metadata.json
      vanilla.arg
      fabric.arg
  assets/
    indexes/
    objects/

Explanation:

- **db.sqlite**  
  SQLite database tracking installed versions and file references.

- **versions/**  
  Per-version directories containing metadata and files required to launch.

- **assets/**  
  Minecraft asset storage following the official structure.

---

## Commands

### update

Downloads the latest version manifest and updates the local database.

This command **does not install any game files**.

---

### list

Lists versions known to the local database.

Options:

- `--release`
- `--snapshot`
- `--installed`

Pattern is an optional glob filter.

Examples:

```sh
gmcl list
gmcl list --release
gmcl list "1.20*"
```

---

### install

Installs a specific Minecraft version.

Steps:

1. Verify the version exists in the database.
2. Download the version metadata.
3. Parse asset and library dependencies.
4. Download missing files.
5. Record file references in SQLite.

Downloads are performed using **32 concurrent workers**.

---

### remove

Removes an installed version.

Shared files used by other versions are preserved.

---

### start

Launches Minecraft using the system Java runtime.

Version resolution:

1. Use explicitly specified version
2. Otherwise use the last started version
3. Otherwise use the latest installed version

gmcl generates the correct Java launch command and executes it.

---

## Fabric Mode

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

## Java Runtime

gmcl does not manage Java installations.

It assumes that `java` is available in the system PATH.

Users are responsible for installing and managing their Java runtime.
