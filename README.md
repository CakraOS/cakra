# CakraOS

> An independent Linux distribution built from scratch, without Linux From Scratch (LFS).

CakraOS adalah proyek distro Linux independen yang dibangun dari nol dengan fokus pada core system, package management, security, dan developer tooling.

Proyek ini dimulai dengan sumber daya minimal dan dikembangkan secara bertahap menggunakan Go sebagai bahasa utama untuk system tooling.

GUI belum menjadi prioritas. Fokus awal CakraOS adalah membangun fondasi sistem yang kuat, modular, aman, dan dapat dikembangkan menjadi ekosistem Linux mandiri.

---

## Vision

CakraOS bertujuan menjadi sistem operasi Linux independen dengan:

- Independent package ecosystem
- Independent package format
- Secure package signing
- Native package manager
- Dependency resolver
- Reproducible build system
- Minimal core system
- System services written primarily in Go
- Developer-friendly tooling
- Future multi-device ecosystem
- Future graphical environment
- No dependency on Linux From Scratch as the project methodology

---

# Project Roadmap

## Phase 0 — Project Foundation

- [x] Define CakraOS concept
- [x] Choose project name: **CakraOS**
- [x] Create GitHub organization
- [x] Create main repository
- [x] Define independent Linux distribution architecture
- [x] Decide to focus on core/headless system first
- [x] Choose Go as the primary system tooling language
- [ ] Define complete system architecture
- [ ] Define filesystem hierarchy
- [ ] Define system directories
- [ ] Define system initialization architecture
- [ ] Define configuration architecture

---

# Phase 1 — Bootstrap & Build System

Goal: create the first native CakraOS package build infrastructure.

- [x] Create `cakra` CLI
- [x] Create package build command
- [x] Create package source structure
- [x] Implement package build staging
- [x] Implement `DESTDIR`
- [x] Generate package manifest
- [x] Build first `hello` package
- [x] Verify staged filesystem
- [x] Fix build environment/path handling
- [ ] Build multiple native packages
- [ ] Define package recipe format
- [ ] Implement build dependencies
- [ ] Implement build isolation
- [ ] Implement reproducible builds
- [ ] Implement build cache
- [ ] Implement parallel builds

### First successful package

```text
hello
└── usr/bin/hello
```

---

# Phase 2 — GPK Package Format

Goal: create CakraOS's independent package format.

## GPK v0.1

- [x] Design GPK format
- [x] Define GPK magic
- [x] Define binary container structure
- [x] Define metadata section
- [x] Define manifest section
- [x] Define payload section
- [x] Use uint64 little-endian section lengths
- [x] Implement GPK writer
- [x] Implement GPK reader
- [x] Implement metadata
- [x] Implement manifest
- [x] Implement payload
- [x] Compress payload using Zstandard
- [x] Generate SHA-256 checksum
- [x] Verify package integrity

Current conceptual format:

```text
GPK
├── MAGIC
├── METADATA
├── MANIFEST
└── PAYLOAD
       └── tar.zst
```

---

# Phase 3 — Package Security

Goal: ensure packages can be authenticated and verified.

- [x] Generate Cakra signing key
- [x] Generate Ed25519 signatures
- [x] Implement signature verification
- [x] Implement package integrity verification
- [x] Detect modified/tampered packages
- [x] Implement `pkg-verify`
- [x] Test tampered GPK
- [ ] Define official key distribution mechanism
- [ ] Define key rotation mechanism
- [ ] Define trusted repository keys
- [ ] Define package trust policy
- [ ] Define key revocation mechanism

Security principle:

```text
Package
   │
   ├── SHA-256
   │
   └── Ed25519
          │
          ▼
       VERIFY
          │
     ┌────┴────┐
     │         │
   VALID     INVALID
     │         │
     ▼         ▼
  INSTALL    REJECT
```

> Private signing keys must never be committed to GitHub.

---

# Phase 4 — Package Installation

Goal: transform GPK into an installed package.

- [x] Implement GPK verification
- [x] Implement payload extraction
- [x] Implement `pkg-install`
- [x] Install files into rootfs
- [x] Validate package before installation
- [x] Implement package metadata storage
- [x] Implement installed package database
- [x] Implement `pkg-list`
- [x] Implement `pkg-files`

Current package database:

```text
var/lib/cakra/
└── packages/
    └── hello/
        └── metadata.json
```

---

# Phase 5 — Package Ownership & Conflict Detection

Goal: prevent packages from corrupting each other's files.

- [x] Track package-owned files
- [x] Implement package ownership lookup
- [x] Implement `db.Owner()`
- [x] Implement conflict detection
- [x] Reject conflicting files
- [x] Allow reinstall of the same package
- [x] Test ownership
- [x] Test package conflict

Example:

```text
hello
└── usr/bin/hello

foo
└── usr/bin/foo
```

Allowed:

```text
hello → usr/bin/hello
foo   → usr/bin/foo
```

Conflict:

```text
hello → usr/bin/hello
foo   → usr/bin/hello
             ^
             └── CONFLICT
```

---

# Phase 6 — Package Removal

Goal: safely remove installed packages.

- [x] Implement `pkg-remove`
- [x] Read package ownership database
- [x] Verify file ownership before deletion
- [x] Remove owned files
- [x] Remove package database entry
- [x] Handle missing files safely
- [x] Test reinstall after removal
- [ ] Implement empty-directory cleanup
- [ ] Implement dependency-aware removal
- [ ] Implement remove confirmation
- [ ] Implement removal transaction

Current lifecycle:

```text
build
  ↓
GPK
  ↓
verify
  ↓
install
  ↓
database
  ↓
remove
```

---

# Phase 7 — Transaction Engine v0.1

Goal: prevent failed package extraction from directly corrupting rootfs.

### Current milestone

- [ ] Integrate transaction engine into `pkg-install`
- [ ] Create transaction ID
- [ ] Create temporary transaction directory
- [ ] Extract package into staging
- [ ] Validate staging
- [ ] Commit staging to rootfs
- [ ] Save package database only after successful commit
- [ ] Cleanup transaction directory
- [ ] Test failed extraction
- [ ] Verify rootfs remains untouched after extraction failure

Target:

```text
GPK
 │
 ▼
VERIFY
 │
 ▼
CONFLICT CHECK
 │
 ▼
TRANSACTION
 │
 ├── staging
 │
 ├── extraction
 │
 ├── validation
 │
 └── commit
 │
 ▼
ROOTFS
 │
 ▼
PACKAGE DATABASE
```

Transaction staging:

```text
/tmp/cakra/
└── cakra-txn-xxxx/
    └── rootfs/
```

---

# Phase 8 — Transaction Engine v0.2

Goal: make installation recoverable and transactional.

- [ ] Implement transaction journal
- [ ] Track created files
- [ ] Track replaced files
- [ ] Track deleted files
- [ ] Implement commit state
- [ ] Implement rollback
- [ ] Recover interrupted transactions
- [ ] Detect incomplete transactions
- [ ] Implement transaction cleanup
- [ ] Test forced installation failure
- [ ] Test rollback
- [ ] Test interrupted transaction recovery

Target:

```text
PREPARE
   ↓
STAGE
   ↓
VALIDATE
   ↓
COMMIT
   │
   ├── SUCCESS
   │
   └── FAILURE
          ↓
       ROLLBACK
```

---

# Phase 9 — Dependency Metadata

Goal: allow packages to declare dependencies.

Example:

```json
{
  "name": "hello",
  "version": "0.1.0",
  "release": 1,
  "architecture": "aarch64",
  "depends": [
    "libfoo >= 1.0",
    "libbar"
  ]
}
```

Tasks:

- [ ] Define dependency syntax
- [ ] Define version constraints
- [ ] Define optional dependencies
- [ ] Define conflicts
- [ ] Define provides
- [ ] Define replaces
- [ ] Add dependencies to GPK metadata
- [ ] Validate dependency metadata
- [ ] Implement dependency comparison

---

# Phase 10 — Dependency Resolver

Goal: automatically resolve package dependencies.

- [ ] Detect missing dependencies
- [ ] Detect installed dependencies
- [ ] Resolve package versions
- [ ] Build dependency graph
- [ ] Detect dependency cycles
- [ ] Detect impossible dependencies
- [ ] Determine installation order
- [ ] Determine removal dependencies
- [ ] Implement dependency transaction

Example:

```text
hello
 ├── libfoo
 │    └── libc
 │
 └── libbar
      └── libc
```

Resolver:

```text
libc
 ↓
libfoo
 ↓
libbar
 ↓
hello
```

---

# Phase 11 — Package Upgrade

Goal: safely upgrade installed packages.

- [ ] Detect installed package version
- [ ] Compare package versions
- [ ] Compare releases
- [ ] Prepare upgrade transaction
- [ ] Preserve configuration files
- [ ] Replace package files safely
- [ ] Update package database
- [ ] Rollback failed upgrade
- [ ] Implement downgrade
- [ ] Implement `pkg-upgrade`

Lifecycle:

```text
old package
     │
     ▼
dependency check
     │
     ▼
transaction
     │
     ▼
new package
     │
     ▼
database update
```

---

# Phase 12 — Package Repository

Goal: create the first CakraOS package repository.

Repository structure:

```text
repository/
├── index.json
├── packages/
│   ├── hello/
│   │   └── hello-0.1.0-1-aarch64.gpk
│   ├── libc/
│   └── ...
└── keys/
    └── cakra-public.key
```

Tasks:

- [ ] Define repository format
- [ ] Define repository metadata
- [ ] Define package index
- [ ] Define repository signing
- [ ] Generate repository index
- [ ] Verify repository index
- [ ] Add package search
- [ ] Add package version lookup
- [ ] Add package download
- [ ] Add repository configuration

---

# Phase 13 — Remote Package Manager

Goal:

```bash
cakra install hello
```

instead of:

```bash
cakra pkg-install hello.gpk
```

Tasks:

- [ ] Repository configuration
- [ ] HTTP/HTTPS client
- [ ] Repository discovery
- [ ] Package search
- [ ] Package download
- [ ] Download checksum verification
- [ ] Signature verification
- [ ] Dependency resolution
- [ ] Transaction installation
- [ ] Cache downloaded packages
- [ ] Offline installation

Target:

```text
cakra install hello
       │
       ▼
repository
       │
       ▼
dependency resolver
       │
       ▼
download
       │
       ▼
verify
       │
       ▼
transaction
       │
       ▼
installed
```

---

# Phase 14 — CakraOS Bootstrap System

Goal: produce a minimal bootable CakraOS userspace.

- [ ] Define base system
- [ ] Build libc
- [ ] Build compiler toolchain
- [ ] Build core utilities
- [ ] Build shell
- [ ] Build filesystem tools
- [ ] Build init system
- [ ] Build logging service
- [ ] Build device management
- [ ] Build user management
- [ ] Build service manager
- [ ] Build networking userspace
- [ ] Create minimal root filesystem
- [ ] Boot first CakraOS userspace

Target:

```text
CakraOS Minimal
├── /bin
├── /sbin
├── /usr
├── /etc
├── /var
├── /home
├── /tmp
├── /dev
├── /proc
└── /sys
```

---

# Phase 15 — Cakra System Core

Goal: create native CakraOS system services.

- [ ] Cakra init
- [ ] Cakra service manager
- [ ] Cakra runtime
- [ ] Cakra system manager
- [ ] Cakra user manager
- [ ] Cakra logging
- [ ] Cakra configuration system
- [ ] Cakra hardware abstraction
- [ ] Cakra application manager
- [ ] Cakra update manager

Go becomes the primary language for CakraOS system tooling.

---

# Phase 16 — Developer SDK

Goal: make application and package development easy.

- [ ] Cakra SDK
- [ ] Package development tools
- [ ] Package template generator
- [ ] Build environment
- [ ] Cross compilation
- [ ] Debugging tools
- [ ] API documentation
- [ ] Developer documentation
- [ ] Application packaging tools
- [ ] SDK version management

Example:

```bash
cakra create package myapp
cakra build myapp
cakra package myapp
cakra install myapp
```

---

# Phase 17 — Security Hardening

- [ ] Secure boot strategy
- [ ] Package trust model
- [ ] Key rotation
- [ ] Key revocation
- [ ] Sandboxing
- [ ] Capability model
- [ ] Application permissions
- [ ] Secure update mechanism
- [ ] Reproducible builds
- [ ] Security audit tooling

---

# Phase 18 — Networking & System Services

- [ ] Network manager
- [ ] DNS management
- [ ] Time synchronization
- [ ] Firewall integration
- [ ] Remote administration
- [ ] System monitoring
- [ ] Resource management
- [ ] Service discovery

---

# Phase 19 — GUI / Desktop

> GUI is intentionally postponed until the core system is stable.

Future direction:

```text
Cakra Core
    │
    ▼
Wayland
    │
    ▼
UI Toolkit
    │
    ▼
Cakra Shell
    │
    ├── Desktop
    ├── Mobile
    └── Embedded
```

Potential technologies will be evaluated later.

Tasks:

- [ ] Wayland integration
- [ ] UI toolkit selection
- [ ] Display server/session architecture
- [ ] Cakra compositor/shell
- [ ] Window management
- [ ] Settings
- [ ] Notification system
- [ ] Application launcher
- [ ] Desktop environment

---

# Phase 20 — CakraOS Multi-Device Ecosystem

Long-term goal:

```text
             CakraOS
                │
       ┌────────┼────────┐
       │        │        │
     Desktop  Mobile   Embedded
       │        │        │
       └────────┼────────┘
                │
          Shared Services
```

Potential targets:

- [ ] Desktop
- [ ] Laptop
- [ ] Smartphone
- [ ] Tablet
- [ ] TV
- [ ] STB
- [ ] Embedded devices
- [ ] Server

---

# Phase 21 — CakraOS Ecosystem

- [ ] Official package repository
- [ ] Developer portal
- [ ] Documentation portal
- [ ] Package submission system
- [ ] Build infrastructure
- [ ] CI/CD
- [ ] Release infrastructure
- [ ] Security infrastructure
- [ ] Community infrastructure
- [ ] Application ecosystem

---

# Final Goal

CakraOS should eventually become a complete independent Linux ecosystem:

```text
                         CakraOS
                            │
              ┌─────────────┼─────────────┐
              │             │             │
           Core          Packages       Security
              │             │             │
           Services      GPK Format    Signing
              │             │             │
           Runtime       Repository      Trust
              │             │             │
              └─────────────┼─────────────┘
                            │
                         SDK / API
                            │
                  ┌─────────┼─────────┐
                  │         │         │
               Desktop    Mobile   Embedded
```

---

# Current Progress

## Current milestone

**Transaction Engine v0.1**

### Completed

- [x] Cakra project initialized
- [x] GitHub organization created
- [x] Go-based Cakra CLI
- [x] Native package build system
- [x] `DESTDIR` staging
- [x] First `hello` package
- [x] GPK v0.1
- [x] Metadata
- [x] Manifest
- [x] Zstandard payload
- [x] SHA-256 verification
- [x] Ed25519 signing
- [x] Tamper detection
- [x] `pkg-verify`
- [x] `pkg-install`
- [x] Package database
- [x] `pkg-list`
- [x] `pkg-files`
- [x] File ownership
- [x] Conflict detection
- [x] `pkg-remove`
- [x] Reinstall package
- [x] Unit tests passing

### Currently working on

- [ ] Transaction staging
- [ ] Transaction commit
- [ ] Transaction cleanup
- [ ] Failed transaction handling
- [ ] Rollback testing

---

# Development Philosophy

CakraOS is developed incrementally.

The project prioritizes:

1. **Core before GUI**
2. **Correctness before features**
3. **Security before convenience**
4. **Small components before complex abstractions**
5. **Tested functionality before milestones are marked complete**
6. **Independent infrastructure**
7. **Go for system tooling**
8. **No unnecessary dependency on existing distro build systems**

---

# Development Environment

Initial development is performed using:

- Android
- Termux
- Git
- GitHub
- Go

The project is intentionally being developed without requiring a traditional PC/Linux workstation during the early stages.

---

# Current Status

```text
CakraOS
│
├── Build System              ██████████░░  In Progress
├── GPK                       ████████████  Complete
├── Package Security          ████████████  Complete
├── Package Installation      ████████████  Complete
├── Package Database          ████████████  Complete
├── Ownership/Conflict        ████████████  Complete
├── Package Removal           ████████████  Complete
├── Transaction Engine        ██░░░░░░░░░░  In Progress
├── Dependencies              ░░░░░░░░░░░░  Planned
├── Upgrade System             ░░░░░░░░░░░░  Planned
├── Repository                 ░░░░░░░░░░░░  Planned
├── Bootstrap                  ░░░░░░░░░░░░  Planned
├── System Core                ░░░░░░░░░░░░  Planned
├── SDK                        ░░░░░░░░░░░░  Planned
├── Security Hardening         ░░░░░░░░░░░░  Planned
├── GUI                        ░░░░░░░░░░░░  Future
└── Multi-Device               ░░░░░░░░░░░░  Future
```

---

# License

See [LICENSE](LICENSE).

---

# CakraOS

**Build the core. Build the system. Build the ecosystem.**