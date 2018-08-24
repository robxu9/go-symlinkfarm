package symlinkfarm

// LinkHandler is a function that creates a link titled newname that
// points to oldname.
type LinkHandler func(oldname, newname string) error

type sourceInfo struct {
	Path   string
	Source string
	IsDir  bool
}

type sourceInfoS []sourceInfo

func (s sourceInfoS) Any(f func(sourceInfo) bool) bool {
	for _, v := range s {
		if f(v) {
			return true
		}
	}
	return false
}

func (s sourceInfoS) All(f func(sourceInfo) bool) bool {
	for _, v := range s {
		if !f(v) {
			return false
		}
	}
	return true
}

func (s sourceInfoS) Sources() []string {
	sources := []string{}
	for _, v := range s {
		sources = append(sources, v.Source)
	}
	return sources
}

func sourceInfoIsDir(s sourceInfo) bool {
	return s.IsDir
}
