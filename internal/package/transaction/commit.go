package transaction

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func Commit(
	staging string,
	root string,
) error {
	return copyTree(
		staging,
		root,
	)
}

func copyTree(
	src string,
	dst string,
) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(
			src,
			entry.Name(),
		)

		dstPath := filepath.Join(
			dst,
			entry.Name(),
		)

		info, err := entry.Info()
		if err != nil {
			return err
		}

		if info.IsDir() {
			if err := os.MkdirAll(
				dstPath,
				info.Mode(),
			); err != nil {
				return err
			}

			if err := copyTree(
				srcPath,
				dstPath,
			); err != nil {
				return err
			}

			continue
		}

		if err := copyFile(
			srcPath,
			dstPath,
			info.Mode(),
		); err != nil {
			return err
		}
	}

	return nil
}

func copyFile(
	src string,
	dst string,
	mode os.FileMode,
) error {
	if err := os.MkdirAll(
		filepath.Dir(dst),
		0755,
	); err != nil {
		return err
	}

	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf(
			"open source: %w",
			err,
		)
	}
	defer in.Close()

	out, err := os.OpenFile(
		dst,
		os.O_CREATE|
			os.O_WRONLY|
			os.O_TRUNC,
		mode,
	)
	if err != nil {
		return fmt.Errorf(
			"open destination: %w",
			err,
		)
	}

	if _, err := io.Copy(
		out,
		in,
	); err != nil {
		out.Close()
		return err
	}

	return out.Close()
}
