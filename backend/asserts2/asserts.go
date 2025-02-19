package asserts2

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

/*** Super duper wrapper for testify/asserts, chain calls, handle t reference storage
 */

type assertsImpl struct {
	t *testing.T
}

func (a *assertsImpl) Empty(v interface{}, msgAndParams ...interface{}) Asserts {
	assert.Empty(a.t, v, msgAndParams...)
	return a
}

func (a *assertsImpl) NotEmpty(v interface{}, msgAndParams ...interface{}) Asserts {
	assert.NotEmpty(a.t, v, msgAndParams...)
	return a
}

func (a *assertsImpl) WithT(t *testing.T) Asserts {
	return Get(t)
}

func (a *assertsImpl) NoError(err error, msgAndParams ...interface{}) Asserts {
	assert.NoError(a.t, err, msgAndParams...)
	return a
}

func (a *assertsImpl) Error(err error, msgAndParams ...interface{}) Asserts {
	assert.Error(a.t, err, msgAndParams...)
	return a
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

type Asserts interface {
	Nil(v interface{}, msgAndParams ...interface{}) Asserts
	NotNil(v interface{}, msgAndParams ...interface{}) Asserts
	Equal(expected, actual interface{}, msgAndParams ...interface{}) Asserts
	NotEqual(expected, actual interface{}, msgAndParams ...interface{}) Asserts
	True(value bool, msgAndParams ...interface{}) Asserts
	False(value bool, msgAndParams ...interface{}) Asserts
	NoError(err error, msgAndParams ...interface{}) Asserts
	Error(err error, msgAndParams ...interface{}) Asserts
	WithT(t *testing.T) Asserts
	Empty(v interface{}, msgAndParams ...interface{}) Asserts
	NotEmpty(v interface{}, msgAndParams ...interface{}) Asserts
}

func Get(t *testing.T) Asserts {
	return &assertsImpl{
		t: t,
	}

}
