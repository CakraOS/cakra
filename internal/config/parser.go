package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/CakraOS/cakra/internal/package"
)

func ParsePackage(path string) (*packagefmt.Package, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	pkg := &packagefmt.Package{}

	scanner := bufio.NewScanner(file)

	var section string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if line == "build:" {
			section = "build"
			continue
		}

		if line == "install:" {
			section = "install"
			continue
		}

		if strings.HasPrefix(line, "- ") {
			value := strings.TrimSpace(strings.TrimPrefix(line, "- "))

			switch section {
			case "build":
				pkg.Build = append(pkg.Build, value)
			case "install":
				pkg.Install = append(pkg.Install, value)
			}

			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid line: %s", line)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		value = strings.Trim(value, `"`)

		switch key {
		case "name":
			pkg.Name = value

		case "version":
			pkg.Version = value

		case "release":
			n, err := strconv.Atoi(value)
			if err != nil {
				return nil, err
			}
			pkg.Release = n

		case "architecture":
			pkg.Architecture = value

		case "source":
			pkg.Source = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if pkg.Name == "" {
		return nil, fmt.Errorf("package name is missing")
	}

	if pkg.Version == "" {
		return nil, fmt.Errorf("package version is missing")
	}

	return pkg, nil
}
