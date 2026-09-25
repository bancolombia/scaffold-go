package commands

import (
	"errors"

	"github.com/gabheadz/scaffold-go/internal/generator"
)

// asMainWiringWarning is a small wrapper around errors.As that lets the
// generate-ep / generate-da commands distinguish "best-effort wiring failed but
// generation succeeded" warnings from harder failures.
func asMainWiringWarning(err error, target **generator.MainWiringWarning) bool {
	return errors.As(err, target)
}
