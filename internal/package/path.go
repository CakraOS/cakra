package packagefmt

import (
	"fmt"
	"path/filepath"
	"strings"
)

func securePath(
	root string,
	name string,
) (string, error) {
	if filepath.IsAbs(name) {
		return "", fmt.Errorf(
			"absolute path rejected: %s",
			name,
		)
	}

	clean := filepath.Clean(name)

	if clean == ".." ||
		strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf(
			"path traversal rejected: %s",
			name,
		)
	}

	target := filepath.Join(root, clean)

	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}

	targetAbs, err := filepath.Abs(target)
	if err != nil {
		return "", err
	}

	if targetAbs != rootAbs &&
		!strings.HasPrefix(
			targetAbs,
			rootAbs+string(filepath.Separator),
		) {
		return "", fmt.Errorf(
			"path escapes root: %s",
			name,
		)
	}

	return targetAbs, nil
}
