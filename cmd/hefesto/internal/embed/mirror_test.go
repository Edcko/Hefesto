package embed

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

// sourceConfigDir is the configuration tree that gets embedded. It is four
// levels up from this package: embed → internal → hefesto → cmd → root.
const sourceConfigDir = "../../../../HefestoOpenCode"

// embeddedFiles returns every file the binary ships, keyed by its path
// relative to the configuration root.
func embeddedFiles(t *testing.T) map[string][]byte {
	t.Helper()

	files := make(map[string][]byte)

	err := fs.WalkDir(ConfigFiles, "config", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == "node_modules" {
				return fs.SkipDir
			}
			return nil
		}
		content, err := fs.ReadFile(ConfigFiles, path)
		if err != nil {
			return err
		}
		relative, err := filepath.Rel("config", path)
		if err != nil {
			return err
		}
		files[filepath.ToSlash(relative)] = content
		return nil
	})
	if err != nil {
		t.Fatalf("failed to walk the embedded configuration: %v", err)
	}

	return files
}

// sourceFiles returns every file of the configuration tree, keyed by its path
// relative to the tree root.
func sourceFiles(t *testing.T) map[string][]byte {
	t.Helper()

	files := make(map[string][]byte)

	err := filepath.WalkDir(sourceConfigDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == "node_modules" || d.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		content, err := os.ReadFile(path) //nolint:gosec // G304: the path comes from the repository tree
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(sourceConfigDir, path)
		if err != nil {
			return err
		}
		files[filepath.ToSlash(relative)] = content
		return nil
	})
	if err != nil {
		t.Fatalf("failed to walk the source configuration: %v", err)
	}

	return files
}

func TestEmbeddedConfigurationMatchesTheSourceTree(t *testing.T) {
	embedded := embeddedFiles(t)
	source := sourceFiles(t)

	if len(embedded) == 0 {
		t.Fatal("the embedded configuration is empty")
	}

	for name, content := range source {
		embeddedContent, ok := embedded[name]
		if !ok {
			t.Errorf("%s is in the source tree but not in the binary", name)
			continue
		}
		if !bytes.Equal(content, embeddedContent) {
			t.Errorf("%s differs between the source tree and the binary", name)
		}
	}

	for name := range embedded {
		if _, ok := source[name]; !ok {
			t.Errorf("%s ships in the binary but not in the source tree", name)
		}
	}
}
