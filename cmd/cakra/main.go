package main

import (
	"fmt"
	"os"

	"github.com/CakraOS/cakra/internal/build"
	packagefmt "github.com/CakraOS/cakra/internal/package"
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

		fmt.Println("Files:")

		for _, file := range gpk.Manifest.Files {
			fmt.Printf("  %s\n", file)
		}
	case "help":
		usage()

	default:
		fmt.Printf("cakra: unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}
