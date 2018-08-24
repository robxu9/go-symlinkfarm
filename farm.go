package symlinkfarm

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/juju/loggo"

	"github.com/pkg/errors"
)

var (
	ErrTargetExist           = errors.New("symlinkfarm: target directory exists")
	ErrFileDirectoryConflict = errors.New("symlinkfarm: file vs directory conflict")

	logger = loggo.GetLogger("lib.symlinkfarm")
)

// FarmConfig specifies the configuration settings for creating a farm.
type FarmConfig struct {
	OnConflict ConflictHandler
	Linker     LinkHandler
}

// DefaultFarmConfig specifies the default configuration for creating a farm.
var DefaultFarmConfig = FarmConfig{
	OnConflict: NeverConflict,
	Linker:     os.Symlink,
}

// Create creates a farm at the specified target directory with the specified
// farm configuration and source directories. The target directory must not
// exist before this operation.
func Create(config FarmConfig, targetDir string, sourceDirs ...string) error {
	logger.Debugf("Create called with %+v, %v, %v", config, targetDir, sourceDirs)

	// ensure that the target directory doesn't exist
	if _, err := os.Lstat(targetDir); !os.IsNotExist(err) {
		logger.Debugf("targetDir(%v) exists", targetDir)
		return errors.WithStack(ErrTargetExist)
	}

	// create a map which will contain the files we have in the target and
	// the source directories it came from (also is it a dir??)
	files, err := determineFarm(sourceDirs...)
	if err != nil {
		logger.Debugf("determineFarm failed: %v", err)
		return errors.WithStack(err)
	}
	logger.Debugf("determineFarm returned %+v", files)

	// simplify the farm down into a list of actions we can take
	actions, err := simplifyFarm(files, targetDir, config)
	if err != nil {
		logger.Debugf("simplifyFarm failed: %v", err)
		return errors.WithStack(err)
	}
	logger.Debugf("simplifyFarm returned %+v", actions)

	// create the target directory
	if err = os.Mkdir(targetDir, os.ModePerm); err != nil {
		logger.Debugf("mkdir(%v) failed: %v", targetDir, err)
		return errors.WithStack(err)
	}
	logger.Debugf("mkdir(%v) succeeded", targetDir)

	// and let's go...
	for i, v := range actions {
		logger.Debugf("action(%v): executing %v on %v", i, v.Action, v.Location)
		if err = v.Perform(); err != nil {
			logger.Debugf("action(%v): failed with %v", err)
			return err
		}
	}

	logger.Debugf("Create finished")
	return nil
}

func determineFarm(sourceDirs ...string) (map[string]sourceInfoS, error) {
	// create a map which will contain the files we have in the target and
	// the source directories it came from (also is it a dir??)
	files := map[string]sourceInfoS{}

	// attempt to map by going through every single source directory
	for _, sourceDir := range sourceDirs {
		if err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
			// do we have an error? if so, fail because we're not going to be able to construct
			if err != nil {
				return errors.Wrap(err, path)
			}

			// strip the source directory out to store for the symlink farm
			farmPath := strings.TrimPrefix(path, sourceDir)

			// if it's an empty string, it's the root directory. that's implied.
			if farmPath == "" {
				return nil
			}

			// add it in
			if _, ok := files[farmPath]; !ok {
				files[farmPath] = sourceInfoS{}
			}

			files[farmPath] = append(files[farmPath], sourceInfo{
				Path:   farmPath,
				Source: path,
				IsDir:  info.IsDir(),
			})

			return nil
		}); err != nil {
			return nil, errors.Wrap(err, sourceDir)
		}
	}

	return files, nil
}
