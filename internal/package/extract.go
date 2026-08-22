package packagefmt

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/klauspost/compress/zstd"
)

func ExtractPayload(
	payload []byte,
	root string,
) error {
	reader, err := zstd.NewReader(
		bytes.NewReader(payload),
	)
	if err != nil {
		return fmt.Errorf(
			"open zstd payload: %w",
			err,
		)
	}
	defer reader.Close()

	tarReader := tar.NewReader(reader)

	for {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			return fmt.Errorf(
				"read tar: %w",
				err,
			)
		}

		target, err := securePath(
			root,
			header.Name,
		)
		if err != nil {
			return err
		}

		switch header.Typeflag {

		case tar.TypeDir:

			if err := os.MkdirAll(
				target,
				os.FileMode(header.Mode),
			); err != nil {
				return err
			}

		case tar.TypeReg:

			if err := os.MkdirAll(
				filepath.Dir(target),
				0755,
			); err != nil {
				return err
			}

			file, err := os.OpenFile(
				target,
				os.O_CREATE|
					os.O_WRONLY|
					os.O_TRUNC,
				os.FileMode(header.Mode),
			)
			if err != nil {
				return err
			}

			if _, err := io.Copy(
				file,
				tarReader,
			); err != nil {
				file.Close()
				return err
			}

			if err := file.Close(); err != nil {
				return err
			}

		default:
			return fmt.Errorf(
				"unsupported tar entry: %s",
				header.Name,
			)
		}
	}

	return nil
}
