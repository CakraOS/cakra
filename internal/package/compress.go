package packagefmt

import (
	"fmt"
	"os"
	"os/exec"
)

func CompressZstd(input string, output string) error {
	cmd := exec.Command(
		"zstd",
		"-q",
		"-f",
		input,
		"-o",
		output,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("zstd compression: %w", err)
	}

	return nil
}
