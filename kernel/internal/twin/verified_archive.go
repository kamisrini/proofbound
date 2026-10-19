package twin

import (
	"archive/tar"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/xi2/xz"
)

// VerifiedArchiveEntry identifies one regular file expected after safe archive extraction.
// Directories are implicit; links and other file types are rejected by this implementation.
type VerifiedArchiveEntry struct {
	Path   string
	Mode   string
	SHA256 string
}

// VerifiedPostgresInput is the complete byte-and-inventory input for an isolated replay.
// ArchiveBytes are consumed synchronously and are never retained or mutated.
type VerifiedPostgresInput struct {
	ArchiveBytes         []byte
	ArchiveSHA256        string
	PostgresBinarySHA256 string
	Inventory            []VerifiedArchiveEntry
}

var errUnsafeArchive = errors.New("twin: unsafe postgres archive")

func validateVerifiedPostgresInput(input VerifiedPostgresInput) error {
	if len(input.ArchiveBytes) == 0 {
		return errors.New("twin: verified postgres archive is empty")
	}
	if err := validateSHA256("archive", input.ArchiveSHA256); err != nil {
		return err
	}
	if err := validateSHA256("postgres binary", input.PostgresBinarySHA256); err != nil {
		return err
	}
	if len(input.Inventory) == 0 {
		return errors.New("twin: verified postgres archive inventory is empty")
	}
	previous := ""
	seen := make(map[string]struct{}, len(input.Inventory))
	for i, entry := range input.Inventory {
		if err := validateArchivePath(entry.Path); err != nil {
			return fmt.Errorf("twin: inventory entry %d: %w", i, err)
		}
		if entry.Path <= previous {
			return fmt.Errorf("twin: postgres archive inventory is not sorted and unique at %q", entry.Path)
		}
		previous = entry.Path
		if _, ok := seen[entry.Path]; ok {
			return fmt.Errorf("twin: postgres archive inventory duplicates %q", entry.Path)
		}
		seen[entry.Path] = struct{}{}
		if entry.Mode != "100644" && entry.Mode != "100755" {
			return fmt.Errorf("twin: postgres archive entry %q has unsupported mode %q", entry.Path, entry.Mode)
		}
		if err := validateSHA256("postgres archive entry "+entry.Path, entry.SHA256); err != nil {
			return err
		}
	}
	for _, required := range []string{"bin/pg_ctl", "bin/postgres"} {
		if _, ok := seen[required]; !ok {
			return fmt.Errorf("twin: postgres archive inventory is missing required %s", required)
		}
	}
	if entry := findArchiveEntry(input.Inventory, "bin/pg_ctl"); entry.Mode != "100755" {
		return errors.New("twin: bin/pg_ctl must be executable")
	}
	if entry := findArchiveEntry(input.Inventory, "bin/postgres"); entry.Mode != "100755" {
		return errors.New("twin: bin/postgres must be executable")
	}
	return nil
}

func validateSHA256(label, value string) error {
	if len(value) != sha256.Size*2 || value != strings.ToLower(value) {
		return fmt.Errorf("twin: %s digest must be lowercase SHA-256", label)
	}
	if _, err := hex.DecodeString(value); err != nil {
		return fmt.Errorf("twin: %s digest is invalid: %w", label, err)
	}
	return nil
}

func findArchiveEntry(entries []VerifiedArchiveEntry, name string) VerifiedArchiveEntry {
	index := sort.Search(len(entries), func(i int) bool { return entries[i].Path >= name })
	if index < len(entries) && entries[index].Path == name {
		return entries[index]
	}
	return VerifiedArchiveEntry{}
}

func stageVerifiedPostgres(ctx context.Context, root string, input VerifiedPostgresInput) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	archiveDigest := sha256.Sum256(input.ArchiveBytes)
	if got := hex.EncodeToString(archiveDigest[:]); got != input.ArchiveSHA256 {
		return "", fmt.Errorf("twin: postgres archive digest mismatch: got %s want %s", got, input.ArchiveSHA256)
	}
	binariesDir := filepath.Join(root, "pgbin")
	if err := os.Mkdir(binariesDir, 0o700); err != nil {
		return "", fmt.Errorf("create binaries directory: %w", err)
	}
	if err := extractVerifiedArchive(ctx, input.ArchiveBytes, binariesDir); err != nil {
		return "", err
	}
	if err := verifyStagedPostgres(binariesDir, input); err != nil {
		return "", err
	}
	return binariesDir, nil
}

func extractVerifiedArchive(ctx context.Context, archiveBytes []byte, destination string) error {
	xzReader, err := xz.NewReader(bytes.NewReader(archiveBytes), 0)
	if err != nil {
		return fmt.Errorf("twin: open postgres txz archive: %w", err)
	}
	tarReader := tar.NewReader(xzReader)
	seen := make(map[string]struct{})
	for {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("twin: extract postgres archive: %w", err)
		}
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("twin: read postgres archive: %w", err)
		}
		name, err := normalizeArchivePath(header.Name)
		if err != nil {
			return err
		}
		target := filepath.Join(destination, filepath.FromSlash(name))
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return fmt.Errorf("twin: create archive directory %q: %w", name, err)
			}
		case tar.TypeReg:
			if _, ok := seen[name]; ok {
				return fmt.Errorf("twin: duplicate postgres archive entry %q", name)
			}
			seen[name] = struct{}{}
			mode := header.Mode & 0o777
			if mode != 0o644 && mode != 0o755 {
				return fmt.Errorf("twin: postgres archive entry %q has unsupported mode %o", name, mode)
			}
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return fmt.Errorf("twin: create parent for archive entry %q: %w", name, err)
			}
			file, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, os.FileMode(mode))
			if err != nil {
				return fmt.Errorf("twin: create archive entry %q: %w", name, err)
			}
			_, copyErr := io.Copy(file, contextReader{ctx: ctx, reader: tarReader})
			closeErr := file.Close()
			if copyErr != nil {
				return fmt.Errorf("twin: extract archive entry %q: %w", name, copyErr)
			}
			if closeErr != nil {
				return fmt.Errorf("twin: close archive entry %q: %w", name, closeErr)
			}
			if err := os.Chmod(target, os.FileMode(mode)); err != nil {
				return fmt.Errorf("twin: set archive entry mode %q: %w", name, err)
			}
		default:
			return fmt.Errorf("%w: archive entry %q has unsupported type %d", errUnsafeArchive, name, header.Typeflag)
		}
	}
	return nil
}

type contextReader struct {
	ctx    context.Context
	reader io.Reader
}

func (r contextReader) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.reader.Read(p)
}

func normalizeArchivePath(name string) (string, error) {
	if name == "" || strings.Contains(name, "\\") || strings.HasPrefix(name, "/") {
		return "", fmt.Errorf("%w: invalid path %q", errUnsafeArchive, name)
	}
	trimmed := strings.TrimSuffix(name, "/")
	clean := path.Clean(trimmed)
	if trimmed == "" || clean == "." || clean != trimmed || clean == ".." || strings.HasPrefix(clean, "../") {
		return "", fmt.Errorf("%w: invalid path %q", errUnsafeArchive, name)
	}
	return clean, nil
}

func validateArchivePath(name string) error {
	_, err := normalizeArchivePath(name)
	return err
}

func verifyStagedPostgres(binariesDir string, input VerifiedPostgresInput) error {
	want := make(map[string]VerifiedArchiveEntry, len(input.Inventory))
	for _, entry := range input.Inventory {
		want[entry.Path] = entry
	}
	seen := make(map[string]struct{}, len(want))
	err := filepath.WalkDir(binariesDir, func(pathName string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(binariesDir, pathName)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if entry.Type()&os.ModeSymlink != 0 || !entry.Type().IsRegular() {
			return fmt.Errorf("%w: staged entry %q is not a regular file", errUnsafeArchive, rel)
		}
		declared, ok := want[rel]
		if !ok {
			return fmt.Errorf("twin: staged postgres archive has unexpected file %q", rel)
		}
		seen[rel] = struct{}{}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		mode := formatArchiveMode(info.Mode().Perm())
		if mode != declared.Mode {
			return fmt.Errorf("twin: staged postgres entry %q mode %s does not match %s", rel, mode, declared.Mode)
		}
		digest, err := fileSHA256(pathName)
		if err != nil {
			return fmt.Errorf("twin: hash staged postgres entry %q: %w", rel, err)
		}
		if digest != declared.SHA256 {
			return fmt.Errorf("twin: staged postgres entry %q digest mismatch", rel)
		}
		return nil
	})
	if err != nil {
		return err
	}
	if len(seen) != len(want) {
		return errors.New("twin: staged postgres archive is missing an inventory entry")
	}
	postgresDigest, err := fileSHA256(filepath.Join(binariesDir, "bin", "postgres"))
	if err != nil {
		return fmt.Errorf("twin: hash launched postgres executable: %w", err)
	}
	if postgresDigest != input.PostgresBinarySHA256 {
		return fmt.Errorf("twin: launched postgres executable digest mismatch: got %s want %s", postgresDigest, input.PostgresBinarySHA256)
	}
	return nil
}

func formatArchiveMode(mode os.FileMode) string {
	switch mode & 0o777 {
	case 0o644:
		return "100644"
	case 0o755:
		return "100755"
	default:
		return ""
	}
}

func fileSHA256(pathName string) (string, error) {
	file, err := os.Open(pathName)
	if err != nil {
		return "", err
	}
	defer func() { _ = file.Close() }()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
