package vault

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilter_Match_NilFilter(t *testing.T) {
	var f *Filter
	assert.True(t, f.Match("any/path"))
}

func TestFilter_Match_NoPrefixes_AllowsAll(t *testing.T) {
	f := NewFilter(nil, nil)
	assert.True(t, f.Match("secrets/foo"))
	assert.True(t, f.Match("config/bar"))
}

func TestFilter_Match_WithPrefix_AllowsMatch(t *testing.T) {
	f := NewFilter([]string{"secrets/"}, nil)
	assert.True(t, f.Match("secrets/foo"))
	assert.False(t, f.Match("config/bar"))
}

func TestFilter_Match_WithMultiplePrefixes(t *testing.T) {
	f := NewFilter([]string{"secrets/", "config/"}, nil)
	assert.True(t, f.Match("secrets/foo"))
	assert.True(t, f.Match("config/bar"))
	assert.False(t, f.Match("other/baz"))
}

func TestFilter_Match_ExcludeTakesPriority(t *testing.T) {
	f := NewFilter([]string{"secrets/"}, []string{"secrets/internal/"})
	assert.True(t, f.Match("secrets/public"))
	assert.False(t, f.Match("secrets/internal/key"))
}

func TestFilter_Match_ExcludeWithNoPrefixes(t *testing.T) {
	f := NewFilter(nil, []string{"tmp/"})
	assert.True(t, f.Match("secrets/foo"))
	assert.False(t, f.Match("tmp/scratch"))
}

func TestFilter_Apply_FiltersSlice(t *testing.T) {
	f := NewFilter([]string{"app/"}, []string{"app/debug/"})
	input := []string{
		"app/config",
		"app/debug/trace",
		"infra/network",
		"app/secrets",
	}
	got := f.Apply(input)
	expected := []string{"app/config", "app/secrets"}
	assert.Equal(t, expected, got)
}

func TestFilter_Apply_EmptyInput(t *testing.T) {
	f := NewFilter([]string{"app/"}, nil)
	got := f.Apply([]string{})
	assert.Empty(t, got)
}

func TestFilter_Apply_NilFilter_ReturnsAll(t *testing.T) {
	var f *Filter
	input := []string{"a", "b", "c"}
	got := f.Apply(input)
	assert.Equal(t, input, got)
}
