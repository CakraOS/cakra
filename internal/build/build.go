package build

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/CakraOS/cakra/internal/config"
	packagefmt "github.com/CakraOS/cakra/internal/package"
)

func runCommand(command string, dir string, destDir string) error {
	fmt.Printf("==> %s\n", command)

	cmd := exec.Command("sh", "-c", command)
	cmd.Dir = dir

	cmd.Env = append(
		os.Environ(),
		"DESTDIR="+destDir,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func Run(packageDir string) error {
	definition := filepath.Join(packageDir, "package.yaml")

	pkg, err := config.ParsePackage(definition)
	if err != nil {
		return err
	}

	fmt.Printf("==> Building %s %s\n", pkg.Name, pkg.Version)
	rootDir, err := os.Getwd()
	if err != nil {
		return err
	}

	buildDir := filepath.Join(rootDir, "build", pkg.Name)
	destDir := filepath.Join(buildDir, "dest")
	//	buildDir := filepath.Join("build", pkg.Name)
	//	destDir := filepath.Join(buildDir, "dest")

	if err := os.RemoveAll(buildDir); err != nil {
		return err
	}

	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return err
	}

	sourceDir := filepath.Join(rootDir, packageDir, pkg.Source)

	// Build phase
	for _, command := range pkg.Build {
		if err := runCommand(command, sourceDir, destDir); err != nil {
			return fmt.Errorf("build failed: %w", err)
		}
	}

	// Install phase
	for _, command := range pkg.Install {
		if err := runCommand(command, sourceDir, destDir); err != nil {
			return fmt.Errorf("install failed: %w", err)
		}
	}

	fmt.Println("==> Staging complete")
	fmt.Println(destDir)
	manifest, err := packagefmt.GenerateManifest(destDir)
	if err != nil {
		return fmt.Errorf("generate manifest: %w", err)
	}

	manifestPath := filepath.Join(buildDir, "manifest")

	if err := manifest.Write(manifestPath); err != nil {
		return fmt.Errorf("write manifest: %w", err)
	}

	fmt.Printf("==> Manifest: %s\n", manifestPath)
	metadata := packagefmt.Metadata{
		Format:       1,
		Name:         pkg.Name,
		Version:      pkg.Version,
		Release:      pkg.Release,
		Architecture: pkg.Architecture,
	}

	outputDir := filepath.Join(rootDir, "output")

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return err
	}

	gpkPath := filepath.Join(
		outputDir,
		fmt.Sprintf(
			"%s-%s-%d-%s.gpk",
			pkg.Name,
			pkg.Version,
			pkg.Release,
			pkg.Architecture,
		),
	)

	if err := packagefmt.BuildGPK(
		gpkPath,
		destDir,
		metadata,
	); err != nil {
		return fmt.Errorf("build gpk: %w", err)
	}

	fmt.Printf("==> GPK: %s\n", gpkPath)

	return nil
}
