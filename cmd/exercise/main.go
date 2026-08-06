// Command exercise reports progress through the single Commerce-to-Router
// journey.
//
// \tgo run ./cmd/exercise                     # journey status
// \tgo run ./cmd/exercise -v                  # status with exact commands
// \tgo run ./cmd/exercise next                # what to work on now
// \tgo run ./cmd/exercise check               # non-zero until complete
// \tgo run ./cmd/exercise optional            # refresh skippable extension evidence
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	verbose := flag.Bool("v", false, "show the focused command for every gate")
	flag.Parse()

	root, err := moduleRoot()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	exitOn(runLearningJourney(root, flag.Arg(0), *verbose))
}

func exitOn(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// moduleRoot walks up from the working directory until it finds go.mod.
func moduleRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no go.mod found above the working directory")
		}
		dir = parent
	}
}
