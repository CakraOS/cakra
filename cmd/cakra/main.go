package main

import (
	"fmt"
	"os"

	"github.com/CakraOS/cakra/internal/build"
	packagefmt "github.com/CakraOS/cakra/internal/package"
	"github.com/CakraOS/cakra/internal/package/db"
)

const (
	name    = "CakraOS"
	version = "0.1.0"
)

func usage() {
	fmt.Println("CakraOS Core")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  cakra version")
	fmt.Println("  cakra build <package>")
	fmt.Println("  cakra help")
}

func main() {
	if len(os.Args) < 2 {
		usage()
		return
	}

	switch os.Args[1] {

	case "version":
		fmt.Printf("%s %s\n", name, version)

	case "build":
		if len(os.Args) < 3 {
			fmt.Println("cakra: missing package")
			os.Exit(1)
		}

		if err := build.Run("packages/" + os.Args[2]); err != nil {
			fmt.Printf("cakra build: %v\n", err)
			os.Exit(1)
		}
	case "pkg-files":
		if len(os.Args) < 3 {
			fmt.Println(
				"cakra pkg-files: missing package name",
			)
			os.Exit(1)
		}

		database := db.New("var/lib/cakra")

		pkg, err := database.Load(os.Args[2])
		if err != nil {
			fmt.Printf(
				"cakra pkg-files: %v\n",
				err,
			)
			os.Exit(1)
		}

		for _, file := range pkg.Files {
			fmt.Println(file)
		}
	case "pkg-remove":
		if len(os.Args) < 3 {
			fmt.Println(
				"cakra pkg-remove: missing package name",
			)
			os.Exit(1)
		}

		root := "rootfs"
		dbRoot := "var/lib/cakra"

		if err := packagefmt.Remove(
			os.Args[2],
			root,
			dbRoot,
		); err != nil {
			fmt.Printf(
				"cakra pkg-remove: %v\n",
				err,
			)
			os.Exit(1)
		}

		fmt.Println(
			"OK: package removed",
		)
	case "pkg-info":
		if len(os.Args) < 3 {
			fmt.Println("cakra pkg-info: missing package")
			os.Exit(1)
		}

		gpk, err := packagefmt.ReadGPK(os.Args[2])
		if err != nil {
			fmt.Printf("cakra pkg-info: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Format:       %d\n", gpk.Metadata.Format)
		fmt.Printf("Name:         %s\n", gpk.Metadata.Name)
		fmt.Printf("Version:      %s\n", gpk.Metadata.Version)
		fmt.Printf("Release:      %d\n", gpk.Metadata.Release)
		fmt.Printf("Architecture: %s\n", gpk.Metadata.Architecture)
		if gpk.Metadata.Checksums != nil {
			fmt.Printf(
				"Payload SHA-256:  %s\n",
				gpk.Metadata.Checksums.Payload,
			)

			fmt.Printf(
				"Manifest SHA-256: %s\n",
				gpk.Metadata.Checksums.Manifest,
			)
		}

		if gpk.Metadata.Signature != nil {
			fmt.Printf(
				"Signature:        %s\n",
				gpk.Metadata.Signature.Algorithm,
			)

			fmt.Printf(
				"Key ID:           %s\n",
				gpk.Metadata.Signature.KeyID,
			)
		}

		fmt.Println("Files:")

		for _, file := range gpk.Manifest.Files {
			fmt.Printf("  %s\n", file)
		}
	case "pkg-verify":
		if len(os.Args) < 3 {
			fmt.Println("cakra pkg-verify: missing package")
			os.Exit(1)
		}

		publicKey := "keys/cakra-public.key"

		if err := packagefmt.VerifyGPK(
			os.Args[2],
			publicKey,
		); err != nil {
			fmt.Printf("INVALID: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("OK: package integrity and signature verified")
	case "keygen":
		if err := packagefmt.GenerateKeyPair(
			"keys/cakra-private.key",
			"keys/cakra-public.key",
		); err != nil {
			fmt.Printf("cakra keygen: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Cakra signing key generated")
	/*case "pkg-install":
	if len(os.Args) < 3 {
		fmt.Println(
			"cakra pkg-install: missing package",
		)
		os.Exit(1)
	}

	root := "rootfs"
	publicKey := "keys/cakra-public.key"

	if err := packagefmt.Install(
		os.Args[2],
		root,
		publicKey,
	); err != nil {
		fmt.Printf(
			"cakra pkg-install: %v\n",
			err,
		)
		os.Exit(1)
	}

	fmt.Println(
		"OK: package installed",
	)
	*/
	case "pkg-list":
		database := db.New("var/lib/cakra")

		packages, err := database.List()
		if err != nil {
			fmt.Printf(
				"cakra pkg-list: %v\n",
				err,
			)
			os.Exit(1)
		}

		fmt.Printf(
			"%-16s %-12s %-8s %s\n",
			"NAME",
			"VERSION",
			"RELEASE",
			"ARCH",
		)

		for _, pkg := range packages {
			fmt.Printf(
				"%-16s %-12s %-8d %s\n",
				pkg.Name,
				pkg.Version,
				pkg.Release,
				pkg.Architecture,
			)
		}
	case "pkg-install":
		if len(os.Args) < 3 {
			fmt.Println(
				"cakra pkg-install: missing package",
			)
			os.Exit(1)
		}

		root := "rootfs"
		dbRoot := "var/lib/cakra"
		publicKey := "keys/cakra-public.key"

		if err := packagefmt.Install(
			os.Args[2],
			root,
			publicKey,
			dbRoot,
		); err != nil {
			fmt.Printf(
				"cakra pkg-install: %v\n",
				err,
			)
			os.Exit(1)
		}

		fmt.Println(
			"OK: package installed",
		)
	case "help":
		usage()

	default:
		fmt.Printf("cakra: unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}
