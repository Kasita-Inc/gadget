package net

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Kasita-Inc/gadget/generator"
)

func TestEmptyAddress(t *testing.T) {
	assert := assert.New(t)
	_, err := NewRedisClient("", generator.String(20))
	assert.EqualError(err, NewInvalidRedisAddressError("", NewEmptyAddressError()).Error())
}

func TestBadAddress(t *testing.T) {
	assert := assert.New(t)
	_, err := NewRedisClient("asdf", generator.String(20))
	assert.Error(err)
}
