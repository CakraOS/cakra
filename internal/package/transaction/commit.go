package transaction

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

var commitHook func(src, dst string) error

type Journal struct {
	CreatedFiles []string
	CreatedDirs  []string
	Backups      []Backup
}

type Backup struct {
	Original string
	Backup   string
}

func Commit(
	staging string,
	root string,
	backupDir string,
) (*Journal, error) {
	journal := &Journal{}

	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		return journal, fmt.Errorf(
			"create backup directory: %w",
			err,
		)
	}

	if err := copyTree(
		staging,
		root,
		backupDir,
		journal,
	); err != nil {
		return journal, err
	}

	return journal, nil
}

func copyTree(
	src string,
	dst string,
	backupDir string,
	journal *Journal,
) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf(
			"read staging directory %s: %w",
			src,
			err,
		)
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
			return fmt.Errorf(
				"stat %s: %w",
				srcPath,
				err,
			)
		}

		if info.IsDir() {
			if err := ensureDirectory(
				dstPath,
				info.Mode().Perm(),
				journal,
			); err != nil {
				return err
			}

			if err := copyTree(
				srcPath,
				dstPath,
				backupDir,
				journal,
			); err != nil {
				return err
			}

			continue
		}

		if err := copyFile(
			srcPath,
			dstPath,
			info.Mode().Perm(),
			backupDir,
			journal,
		); err != nil {
			return err
		}
	}

	return nil
}

func ensureDirectory(
	path string,
	mode os.FileMode,
	journal *Journal,
) error {
	info, err := os.Stat(path)

	if err == nil {
		if !info.IsDir() {
			return fmt.Errorf(
				"destination exists but is not directory: %s",
				path,
			)
		}

		return nil
	}

	if !os.IsNotExist(err) {
		return fmt.Errorf(
			"check directory %s: %w",
			path,
			err,
		)
	}

	if err := os.MkdirAll(
		path,
		mode,
	); err != nil {
		return fmt.Errorf(
			"create directory %s: %w",
			path,
			err,
		)
	}

	journal.CreatedDirs = append(
		journal.CreatedDirs,
		path,
	)

	return nil
}

func copyFile(
	src string,
	dst string,
	mode os.FileMode,
	backupDir string,
	journal *Journal,
) error {
	if err := os.MkdirAll(
		filepath.Dir(dst),
		0o755,
	); err != nil {
		return fmt.Errorf(
			"create parent directory for %s: %w",
			dst,
			err,
		)
	}

	if _, err := os.Stat(dst); err == nil {
		backupPath := filepath.Join(
			backupDir,
			backupName(dst),
		)

		if err := os.MkdirAll(
			filepath.Dir(backupPath),
			0o755,
		); err != nil {
			return fmt.Errorf(
				"create backup parent: %w",
				err,
			)
		}

		if err := copyRaw(
			dst,
			backupPath,
		); err != nil {
			return fmt.Errorf(
				"backup %s: %w",
				dst,
				err,
			)
		}

		journal.Backups = append(
			journal.Backups,
			Backup{
				Original: dst,
				Backup:   backupPath,
			},
		)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf(
			"check destination %s: %w",
			dst,
			err,
		)
	} else {
		journal.CreatedFiles = append(
			journal.CreatedFiles,
			dst,
		)
	}
	if commitHook != nil {
		if err := commitHook(src, dst); err != nil {
			return err
		}
	}

	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf(
			"open source %s: %w",
			src,
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
			"open destination %s: %w",
			dst,
			err,
		)
	}

	if _, err := io.Copy(
		out,
		in,
	); err != nil {
		out.Close()

		return fmt.Errorf(
			"copy %s -> %s: %w",
			src,
			dst,
			err,
		)
	}

	if err := out.Chmod(mode); err != nil {
		out.Close()

		return fmt.Errorf(
			"chmod %s: %w",
			dst,
			err,
		)
	}

	if err := out.Close(); err != nil {
		return fmt.Errorf(
			"close destination %s: %w",
			dst,
			err,
		)
	}

	return nil
}

func copyRaw(
	src string,
	dst string,
) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	info, err := in.Stat()
	if err != nil {
		return err
	}

	out, err := os.OpenFile(
		dst,
		os.O_CREATE|
			os.O_WRONLY|
			os.O_TRUNC,
		info.Mode().Perm(),
	)
	if err != nil {
		return err
	}

	if _, err := io.Copy(
		out,
		in,
	); err != nil {
		out.Close()
		return err
	}

	if err := out.Chmod(
		info.Mode().Perm(),
	); err != nil {
		out.Close()
		return err
	}

	return out.Close()
}

func backupName(path string) string {
	clean := filepath.Clean(path)

	volume := filepath.VolumeName(clean)

	if volume != "" {
		clean = clean[len(volume):]
	}

	clean = filepath.Clean(clean)

	for len(clean) > 0 &&
		(clean[0] == '/' ||
			clean[0] == '\\') {
		clean = clean[1:]
	}

	return filepath.FromSlash(clean)
}
