package generator

import "os"

type Generator interface {
	Validate() error
	Generate() ([]GeneratedFile, error)
	ComponentType() string
}

type GeneratedFile struct {
	Path    string
	Content string
}

// exists is a small filesystem helper shared by generators that need to skip
// re-rendering files that are already present on disk (e.g. the broker
// foundation module).
func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
