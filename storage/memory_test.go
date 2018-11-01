package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Kasita-Inc/gadget/generator"
)

type Foo struct {
	Bar string
	Baz int
}

func TestNewMemoryStorage(t *testing.T) {
	assert := assert.New(t)
	storage := NewMemoryStorage()
	assert.NotNil(storage)
}

func Test_memoryProvider_ReadWriteExistsDelete(t *testing.T) {
	assert := assert.New(t)
	storage := NewMemoryStorage()
	expected := &Foo{Bar: "awef", Baz: 42}
	actual := &Foo{}
	key := "key"
	assert.EqualError(storage.Read("not the key", actual), NoEntryForKey)
	assert.NoError(storage.Write(key, expected))
	assert.True(storage.Exists(key))
	assert.NoError(storage.Read(key, actual))
	assert.Equal(expected, actual)
	assert.NoError(storage.Delete(key))
	assert.EqualError(storage.Read(key, actual), NoEntryForKey)
	assert.False(storage.Exists(key))
}

func Test_memoryProvider_Shard(t *testing.T) {
	assert := assert.New(t)
	storage := NewMemoryStorage()
	pkey := generator.String(20)
	pvalue := generator.String(20)
	skey := generator.String(20)
	svalue := generator.String(20)
	assert.NoError(storage.Write(pkey, pvalue))
	shardName := generator.String(20)
	shard, err := storage.Shard(shardName)
	assert.NoError(err)
	assert.False(shard.Exists(pkey))
	assert.NoError(shard.Write(skey, svalue))
	assert.False(storage.Exists(skey))
	shard, err = storage.Shard(shardName)
	assert.NoError(err)
	assert.False(shard.Exists(pkey))
	assert.True(shard.Exists(skey))
}
