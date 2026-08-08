package build

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/CakraOS/cakra/internal/config"
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

	return nil
}
