package symlinkfarm

import "errors"

var (
	// ErrFileConflict indicates that there was am unresolvable file conflict
	ErrFileConflict = errors.New("symlinkfarm: unresolvable conflict between two or more files")
)

// ConflictHandler takes a set of source files and returns the
// winning one, or errors out if it cannot amicably resolve them.
type ConflictHandler func(...string) (string, error)

// NeverConflict is a ConflictHandler which will always error.
func NeverConflict(sourceFiles ...string) (string, error) {
	return "", ErrFileConflict
}

// IgnoreBothConflict is a ConflictHandler that simply ignores any conflicts by symlinking nothing.
func IgnoreBothConflict(sourceFiles ...string) (string, error) {
	return "", nil
}
