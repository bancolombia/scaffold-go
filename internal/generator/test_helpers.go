package generator

import (
	"os"
	"strconv"

	"github.com/leanovate/gopter"
)

// GetTestParameters returns gopter test parameters with configurable iterations
// Reads GOPTER_MIN_SUCCESS environment variable, defaults to 100
func GetTestParameters() *gopter.TestParameters {
	params := gopter.DefaultTestParameters()

	// Allow configuration via environment variable
	if minSuccess := os.Getenv("GOPTER_MIN_SUCCESS"); minSuccess != "" {
		if n, err := strconv.Atoi(minSuccess); err == nil && n > 0 {
			params.MinSuccessfulTests = n
		}
	} else {
		// Default to 100 for full test runs
		params.MinSuccessfulTests = 100
	}

	return params
}
