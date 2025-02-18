package asserts2

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type assertsImpl struct {
	t *testing.T
}

func Get(t *testing.T) Asserts {
	return &assertsImpl{
		t: t,
	}

}

type Asserts interface {
	Nil(v interface{}, msgAndParams ...interface{}) Asserts
	NotNil(v interface{}, msgAndParams ...interface{}) Asserts
	Equal(expected, actual interface{}, msgAndParams ...interface{}) Asserts
	NotEqual(expected, actual interface{}, msgAndParams ...interface{}) Asserts
	True(value bool, msgAndParams ...interface{}) Asserts
	False(value bool, msgAndParams ...interface{}) Asserts
}

func (a *assertsImpl) Nil(v interface{}, msgAndParams ...interface{}) Asserts {
	assert.Nil(a.t, v, msgAndParams...)
	return a
}

func (a *assertsImpl) NotNil(v interface{}, msgAndParams ...interface{}) Asserts {
	assert.NotNil(a.t, v, msgAndParams...)
	return a
}

func (a *assertsImpl) Equal(expected, actual interface{}, msgAndParams ...interface{}) Asserts {
	assert.Equal(a.t, expected, actual, msgAndParams...)
	return a
}

func (a *assertsImpl) NotEqual(expected, actual interface{}, msgAndParams ...interface{}) Asserts {
	assert.NotEqual(a.t, expected, actual, msgAndParams...)
	return a
}

func (a *assertsImpl) True(value bool, msgAndParams ...interface{}) Asserts {
	assert.True(a.t, value, msgAndParams...)
	return a
}

func (a *assertsImpl) False(value bool, msgAndParams ...interface{}) Asserts {
	assert.False(a.t, value, msgAndParams...)
	return a
}
