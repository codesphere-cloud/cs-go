## cs install skills

Install the Codesphere skills bundled with this CLI

### Synopsis

Installs the Codesphere skill package(s) bundled with this cs CLI release
into a local skills folder (".agent/skills" by default), for use by AI coding
agents that support the Agent Skills format.

Every installed skill folder gets a version.json recording which cs CLI version
last installed or updated it. Re-running this command compares that installed
version against the running cs CLI's own version and only overwrites a skill
when the CLI is newer - so it's safe to run again after "cs update" to pick up
newer skills, and a no-op otherwise.

```
cs install skills [flags]
```

### Examples

```
# Install/update all bundled skills into ./.agent/skills
$ cs install skills 

# Install/update into ~/.agent/skills instead
$ cs install skills --global

# Install/update into a custom directory
$ cs install skills --dir path/to/skills

# Reinstall even if already up to date
$ cs install skills --force
```

### Options

```
      --dir string   Target skills directory (defaults to ".agent/skills", or "~/.agent/skills" with --global)
      --dry-run      Show what would be installed without writing any files
      --force        Reinstall every skill even if it's already up to date
      --global       Install into the user's home directory instead of the current directory
  -h, --help         help for skills
```

### Options inherited from parent commands

```
  -a, --api string      URL of Codesphere API (can also be CS_API)
  -O, --org string      Organization ID (relevant for some commands)
  -t, --team int        Team ID (relevant for some commands, can also be CS_TEAM_ID) (default -1)
  -v, --verbose         Verbose output
  -w, --workspace int   Workspace ID (relevant for some commands, can also be CS_WORKSPACE_ID) (default -1)
```

### SEE ALSO

* [cs install](cs_install.md)	 - Install optional Codesphere CLI extensions

