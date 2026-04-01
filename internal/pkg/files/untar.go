package files

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
)

// IsArchive returns true if the provided path has a .tar, .tar.gz, or .tgz extension.
func IsArchive(path string) bool {
	return strings.HasSuffix(path, ".tar") || strings.HasSuffix(path, ".tar.gz") || strings.HasSuffix(path, ".tgz")
}

// Option defines a functional option for configuring the Untar function.
type Option func(*untarOptions)

// RemoveArchive is an option that configures whether to remove the archive file after extraction.
func RemoveArchive(remove bool) Option {
	return func(opts *untarOptions) {
		opts.removeArchive = remove
	}
}

type untarOptions struct {
	// removeArchive when set to true, will remove the archive file after extraction.
	removeArchive bool
}

// Untar extracts the contents of a tar or tar.gz archive to the specified destination directory.
func Untar(path string, dest string, options ...Option) error {
	if !filepath.IsAbs(dest) {
		return fmt.Errorf("destination path must be absolute: %s", dest)
	}

	opts := untarOptions{}
	for _, opt := range options {
		opt(&opts)
	}

	r, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open archive file: %w", err)
	}
	defer r.Close()

	lr := io.LimitReader(r, math.MaxInt64)

	if isGzip(path) {
		err = untgz(lr, dest, opts)
	} else {
		err = untar(lr, dest, opts)
	}

	if err != nil {
		return fmt.Errorf("failed to extract archive: %w", err)
	}

	if opts.removeArchive {
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("failed to remove archive file: %w", err)
		}
	}

	return nil

}

func untgz(r io.Reader, dest string, opts untarOptions) error {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	return untar(gzr, dest, opts)
}

func untar(r io.Reader, dest string, _ untarOptions) error {
	tr := tar.NewReader(r)

	dest = filepath.Clean(dest)

L:
	for {
		header, err := tr.Next()
		switch {
		case err == io.EOF:
			break L
		case err != nil:
			return fmt.Errorf("failed to read tar header: %w", err)
		case header == nil:
			fallthrough
		case header.Name == ".":
			fallthrough
		case header.Name == "./":
			continue L
		}

		target, err := sanitizeArchivePath(dest, header.Name)
		if err != nil {
			return err
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, header.FileInfo().Mode()); err != nil {
					return fmt.Errorf("failed to create directory: %w", err)
				}
			}
		case tar.TypeReg:
			err := os.MkdirAll(filepath.Dir(target), 0o750)
			if err != nil {
				return fmt.Errorf("failed to create directory for file '%s': %w", target, err)
			}

			if err := writeFile(target, header.FileInfo().Mode(), tr); err != nil {
				return err
			}
		}
	}
	return nil
}

func writeFile(target string, mode os.FileMode, r io.Reader) error {
	err := os.MkdirAll(filepath.Dir(target), 0o750)
	if err != nil {
		return fmt.Errorf("failed to create directory for file '%s': %w", target, err)
	}

	f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("failed to create file '%s': %w", target, err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	if _, err := io.Copy(w, r); err != nil {
		return fmt.Errorf("failed to write file '%s': %w", target, err)
	}
	if err := w.Flush(); err != nil {
		return fmt.Errorf("failed to flush file '%s': %w", target, err)
	}
	return nil
}

func isGzip(path string) bool {
	return strings.HasSuffix(path, ".gz") || strings.HasSuffix(path, ".tgz")
}

func sanitizeArchivePath(d, t string) (string, error) {
	target := filepath.Join(d, t)
	if !strings.HasPrefix(target, d+string(os.PathSeparator)) {
		return "", fmt.Errorf("%s: illegal file path", t)
	}
	return target, nil
}
