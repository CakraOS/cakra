# CakraOS Architecture

## Vision

CakraOS is an independent Linux distribution designed
to be built, maintained, and eventually self-hosted
using its own tooling.

## Core principles

1. Independent distribution
2. Minimal core
3. Reproducible builds
4. Signed packages
5. Simple system architecture
6. Go-first userspace tooling
7. Upstream compatibility
8. Cross-platform architecture
9. Self-hosting as the long-term goal

## Technology

Kernel:
- Linux

Primary systems language:
- Go

Low-level components:
- C
- Assembly
- upstream toolchains

Package format:
- .gpk

Package manager:
- cakra

Build system:
- cakra build

Repository:
- Cakra Repository

Init:
- Cakra Init

Service supervisor:
- Cakra Service

## Non-goals

CakraOS core will initially not include:

- Desktop environment
- GUI
- Wayland compositor
- EFL
- Mobile UI
- Desktop installer

These components may be implemented later.
