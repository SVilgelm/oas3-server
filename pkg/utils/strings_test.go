package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDoubleMapStringSet(t *testing.T) {
	v := make(DoubleMapString)
	assert.Nil(t, v["foo"])
	assert.Empty(t, v["foo"]["bar"])
	assert.Nil(t, v["foo"])
	v.Set("foo", "bar", "xyz")
	assert.NotNil(t, v["foo"])
	assert.Equal(t, "xyz", v["foo"]["bar"])
}

func TestContains(t *testing.T) {
	ss := []string{"foo", "bar"}
	assert.True(t, Contains(ss, "foo"))
	assert.False(t, Contains(ss, "xyz"))
}
