package symlinkfarm

import (
	"io"
	"os"
	"strings"

	"github.com/juju/loggo"
	"github.com/karrick/godirwalk"

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

	// ensure that the target directory doesn't exist (or if it exists, is empty)
	createTargetDir := true
	if _, err := os.Lstat(targetDir); !os.IsNotExist(err) {
		logger.Debugf("targetDir(%v) exists", targetDir)

		targetDirFile, err := os.Open(targetDir)
		if err != nil {
			logger.Debugf("targetDir(%v) can't open: %v", targetDir, err)
			return errors.WithStack(err)
		}

		if _, err = targetDirFile.Readdirnames(1); err != io.EOF {
			// not an empty directory
			logger.Debugf("targetDir(%v) not empty: %v", targetDir, err)
			return errors.WithStack(ErrTargetExist)
		}

		createTargetDir = false
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

	// create the target directory (if necessary)
	if createTargetDir {
		if err = os.Mkdir(targetDir, os.ModePerm); err != nil {
			logger.Debugf("mkdir(%v) failed: %v", targetDir, err)
			return errors.WithStack(err)
		}
		logger.Debugf("mkdir(%v) succeeded", targetDir)
	} else {
		logger.Debugf("mkdir(%v) not necessary - targetDir existed but was empty")
	}

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
		if err := godirwalk.Walk(sourceDir, &godirwalk.Options{
			Callback: func(path string, de *godirwalk.Dirent) error {
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
					IsDir:  de.IsDir() && !de.IsSymlink(),
				})

				return nil
			},
			Unsorted: true, // we don't care; we have to sort this later anyway
		}); err != nil {
			return nil, errors.Wrap(err, sourceDir)
		}
	}

	return files, nil
}
