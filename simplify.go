package symlinkfarm

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/pkg/errors"
)

func simplifyFarm(uglyFarm map[string]sourceInfoS, targetDir string, farmConfig FarmConfig) ([]FarmAction, error) {

	actions := []FarmAction{}

	// first, go through and eliminate conflicts/duplicates
	for farmPath, sources := range uglyFarm {
		if sources.Any(sourceInfoIsDir) { // directory handling
			if sources.All(sourceInfoIsDir) { // all directories
				// mkdir
				actions = append(actions, createMkdirAction(targetDir, farmPath))
			} else {
				// conflict in file vs directory
				return nil, errors.WithStack(ErrFileDirectoryConflict)
			}
		} else { // handle file conflicts
			source := sources[0].Source
			var err error
			if len(sources) > 1 { // more than one file, handle file conflicts
				source, err = farmConfig.OnConflict(sources.Sources()...)
				if err != nil {
					// fail
					return nil, err
				}
			}
			if source != "" { // some OnConflict handlers can choose to symlink nothing. weird.
				actions = append(actions, createLinkAction(targetDir, source, farmPath, farmConfig.Linker))
			}
		}
	}

	// then, sort into an array so we can create actions in-order
	// (imagine creating symlinks without a corresponding directory. yikes.)
	sort.Slice(actions, func(i, j int) bool {
		return actions[i].Location < actions[j].Location
	})

	// finally return the list of actions to take
	return actions, nil
}

// FarmAction represents the action needed to create a piece of the farm
type FarmAction struct {
	Location string
	Action   string
	Perform  func() error
}

// target will point to source
func createLinkAction(targetDir, source, target string, linker LinkHandler) FarmAction {
	return FarmAction{
		Location: target,
		Action:   "link",
		Perform: func() error {
			return linker(source, filepath.Join(targetDir, target))
		},
	}
}

func createMkdirAction(targetDir, target string) FarmAction {
	return FarmAction{
		Location: target,
		Action:   "mkdir",
		Perform: func() error {
			return os.Mkdir(filepath.Join(targetDir, target), os.ModePerm)
		},
	}
}

// sorting for FarmAction
