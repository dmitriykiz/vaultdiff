package vault

import "strings"

// Filter holds options for filtering secret keys during comparison.
type Filter struct {
	Prefixes []string
	Exclude  []string
}

// NewFilter creates a Filter from include-prefix and exclude-prefix slices.
func NewFilter(prefixes, exclude []string) *Filter {
	return &Filter{
		Prefixes: prefixes,
		Exclude:  exclude,
	}
}

// Match returns true if the given path should be included according to the filter rules.
// A path is included when:
//  1. It matches at least one prefix (or no prefixes are configured), AND
//  2. It does not match any exclude prefix.
func (f *Filter) Match(path string) bool {
	if f == nil {
		return true
	}

	for _, ex := range f.Exclude {
		if strings.HasPrefix(path, ex) {
			return false
		}
	}

	if len(f.Prefixes) == 0 {
		return true
	}

	for _, p := range f.Prefixes {
		if strings.HasPrefix(path, p) {
			return true
		}
	}

	return false
}

// Apply filters a slice of paths, returning only those that match the filter.
func (f *Filter) Apply(paths []string) []string {
	result := make([]string, 0, len(paths))
	for _, p := range paths {
		if f.Match(p) {
			result = append(result, p)
		}
	}
	return result
}
