package symlinkfarm

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCreate(t *testing.T) {
	Convey("given a temporary folder in which to do work", t, func() {
		tmpDir, err := ioutil.TempDir("", "symlinkfarm-testCreate")
		So(err, ShouldBeNil)

		Printf("tmpDir: %v\n", tmpDir)

		Convey("if targetDir already exists", func() {

			Convey("...but is empty...", func() {
				err := Create(DefaultFarmConfig, tmpDir)

				Convey("continue on fine", func() {
					So(err, ShouldBeNil)
				})
			})

			Convey("...but isn't empty...", func() {
				randomFile, err := os.Create(filepath.Join(tmpDir, "random-file"))
				So(err, ShouldBeNil)
				So(randomFile.Close(), ShouldBeNil)

				Convey("error out", func() {
					err = Create(DefaultFarmConfig, tmpDir)
					So(err, ShouldNotBeNil)
					So(errors.Cause(err), ShouldEqual, ErrTargetExist)
				})
			})
		})

		Convey("when we have two directories that should not conflict with each other", func() {
			folders := []string{"tests/a", "tests/b"}

			Convey("call create and target tmpDir/create-test-1", func() {
				targetDir := filepath.Join(tmpDir, "create-test-1")
				Printf("targetDir: %v\n", targetDir)
				err := Create(DefaultFarmConfig, targetDir, folders...)

				Convey("and ensure that there was no error", func() {
					So(err, ShouldBeNil)
				})
			})
		})

		Convey("when we have two directories that conflict in shared files", func() {
			folders := []string{"tests/a", "tests/c"}

			Convey("call create and target tmpDir/create-test-2", func() {
				targetDir := filepath.Join(tmpDir, "create-test-2")
				Printf("targetDir: %v\n", targetDir)
				err := Create(DefaultFarmConfig, targetDir, folders...)

				Convey("and ensure that ErrFileConflict was returned", func() {
					So(err, ShouldNotBeNil)
					So(errors.Cause(err), ShouldEqual, ErrFileConflict)
				})
			})
		})

		Convey("when we have two directories that conflict in file vs directory", func() {
			folders := []string{"tests/c", "tests/d"}

			Convey("call create and target tmpDir/create-test-3", func() {
				targetDir := filepath.Join(tmpDir, "create-test-3")
				Printf("targetDir: %v\n", targetDir)
				err := Create(DefaultFarmConfig, targetDir, folders...)

				Convey("and ensure that ErrFileDirectoryConflict was returned", func() {
					So(err, ShouldNotBeNil)
					So(errors.Cause(err), ShouldEqual, ErrFileDirectoryConflict)
				})
			})
		})

		Reset(func() {
			os.RemoveAll(tmpDir)
		})
	})
}

func TestDetermineFarm(t *testing.T) {

	Convey("given two folders with content that should not conflict", t, func() {
		folders := []string{"tests/a", "tests/b"}

		Convey("when determineFarm is called", func() {
			farmResult, err := determineFarm(folders...)

			Convey("the farm should return a correct result", func() {
				So(err, ShouldBeNil)
				So(farmResult, ShouldResemble, map[string]sourceInfoS{
					"/a_dir": sourceInfoS{
						sourceInfo{
							Path:   "/a_dir",
							Source: "tests/a/a_dir",
							IsDir:  true,
						},
					},
					"/a_dir/a_dir_file": sourceInfoS{
						sourceInfo{
							Path:   "/a_dir/a_dir_file",
							Source: "tests/a/a_dir/a_dir_file",
							IsDir:  false,
						},
					},
					"/shared_dir": sourceInfoS{
						sourceInfo{
							Path:   "/shared_dir",
							Source: "tests/a/shared_dir",
							IsDir:  true,
						},
						sourceInfo{
							Path:   "/shared_dir",
							Source: "tests/b/shared_dir",
							IsDir:  true,
						},
					},
					"/shared_dir/a_shared_file": sourceInfoS{
						sourceInfo{
							Path:   "/shared_dir/a_shared_file",
							Source: "tests/a/shared_dir/a_shared_file",
							IsDir:  false,
						},
					},
					"/shared_dir/b_shared_file": sourceInfoS{
						sourceInfo{
							Path:   "/shared_dir/b_shared_file",
							Source: "tests/b/shared_dir/b_shared_file",
							IsDir:  false,
						},
					},
					"/a_file": sourceInfoS{
						sourceInfo{
							Path:   "/a_file",
							Source: "tests/a/a_file",
							IsDir:  false,
						},
					},
					"/b_file": sourceInfoS{
						sourceInfo{
							Path:   "/b_file",
							Source: "tests/b/b_file",
							IsDir:  false,
						},
					},
					"/c_file": sourceInfoS{
						sourceInfo{
							Path:   "/c_file",
							Source: "tests/a/c_file",
							IsDir:  false,
						},
					},
				})
			})
		})
	})
}
