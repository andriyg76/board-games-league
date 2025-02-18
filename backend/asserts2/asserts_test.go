package asserts2

import (
	"testing"
)

func TestAsserts_Nil(t *testing.T) {
	asserts := Get(t)
	asserts.Nil(nil)
	asserts.Nil(nil, "should be nil")
}

func TestAsserts_NotNil(t *testing.T) {
	asserts := Get(t)
	asserts.NotNil("not nil")
	asserts.NotNil("not nil", "should not be nil")
}

func TestAsserts_Equal(t *testing.T) {
	asserts := Get(t)
	asserts.Equal(1, 1)
	asserts.Equal(1, 1, "should be equal")
}

func TestAsserts_NotEqual(t *testing.T) {
	asserts := Get(t)
	asserts.NotEqual(1, 2)
	asserts.NotEqual(1, 2, "should not be equal")
}

func TestAsserts_True(t *testing.T) {
	asserts := Get(t)
	asserts.True(true)
	asserts.True(true, "should be true")
}

func TestAsserts_False(t *testing.T) {
	asserts := Get(t)
	asserts.False(false)
	asserts.False(false, "should be false")
}
