package get

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

// Complete returns repository paths relative to GETPATH that match the given prefix,
// completing one path segment at a time.
func Complete(prefix string) ([]string, error) {
	trailingSlash := strings.HasSuffix(prefix, "/") || strings.HasSuffix(prefix, string(filepath.Separator))
	if prefix != "" {
		prefix = strings.ToLower(filepath.ToSlash(filepath.Clean(prefix)))
	}
	prefixDepth := strings.Count(prefix, "/")
	if trailingSlash {
		prefixDepth++
	}

	getpath, err := AbsolutePath()
	if err != nil {
		return nil, fmt.Errorf("resolving GETPATH: %w", err)
	}

	var matches []string
	err = filepath.WalkDir(getpath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if path == getpath {
				return err
			}
			return nil
		}
		if !d.IsDir() || path == getpath {
			return nil
		}
		if strings.HasPrefix(d.Name(), ".") {
			return fs.SkipDir
		}

		rel, err := filepath.Rel(getpath, path)
		if err != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		relLower := strings.ToLower(rel)
		relDepth := strings.Count(rel, "/")

		if relDepth == prefixDepth {
			if strings.HasPrefix(relLower, prefix) {
				if !isGitRepository(path) {
					rel += "/"
				}
				matches = append(matches, rel)
			}
			return fs.SkipDir
		}

		if relLower != prefix && !strings.HasPrefix(prefix, relLower+"/") {
			return fs.SkipDir
		}
		return nil
	})
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	return matches, err
}
