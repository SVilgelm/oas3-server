package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDoubleMapBoolSet(t *testing.T) {
	v := make(DoubleMapBool)
	assert.Nil(t, v["foo"])
	assert.False(t, v["foo"]["bar"])
	assert.Nil(t, v["foo"])
	v.Set("foo", "bar", true)
	assert.NotNil(t, v["foo"])
	assert.True(t, v["foo"]["bar"])
}
