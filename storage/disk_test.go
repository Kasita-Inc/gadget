package storage

import (
	"fmt"
	"os"
	"path"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Kasita-Inc/gadget/generator"
)

var RootPath = getRootPath()

const TestRoot = "disk_test"

type TestStruct struct {
	Foo string
	Bar int
}

func TestMain(m *testing.M) {
	res := m.Run()
	cleanup()
	os.Exit(res)
}

func getRootPath() string {
	uuid := generator.String(10)
	rootPath := fmt.Sprintf("/tmp/%s/%s", TestRoot, strings.TrimSpace(uuid))
	return rootPath
}

func cleanup() {
	wd, _ := os.Getwd()
	os.Chdir("/tmp")
	os.RemoveAll(TestRoot)
	os.Chdir(wd)
}

func TestInitialize(t *testing.T) {
	_, err := NewDisk(RootPath, 0777)
	if err != nil {
		t.Error(err.Error())
	}
	// relative path
	_, err = NewDisk("foo/bar", 0777)
	if err == nil {
		t.Error("Relative path should fail.")
	}
	// empty path
	_, err = NewDisk("", 0777)
	if err == nil {
		t.Error("Empty path should fail.")
	}
}

func TestGetFullPath(t *testing.T) {
	provider, _ := NewDisk(RootPath, 0777)
	impl := provider.(*disk)
	// empty key
	_, err := impl.getFullPath("")
	if err == nil {
		t.Error("Empty partition should fail.")
	}

	key := generator.String(20)
	expected := RootPath + "/" + keyToFilename(key) + ".gob"
	actual, err := impl.getFullPath(key)
	if err != nil {
		t.Error(err)
	}
	if path.Clean(actual) != path.Clean(expected) {
		t.Errorf("'%s' =/= '%s'", actual, expected)
	}
}

func TestSymmetricWrite(t *testing.T) {
	expected := &TestStruct{Foo: "Foo", Bar: 20}
	provider, _ := NewDisk(RootPath, 0777)
	key := generator.String(20)
	err := provider.Write(key, expected)
	if err != nil {
		t.Error(err)
	}
	var actual = &TestStruct{}
	err = provider.Read(key, actual)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("%v != %v", actual, expected)
	}
}

func TestExists(t *testing.T) {
	s := &TestStruct{Foo: "foo", Bar: 20}
	provider := &disk{RootPath: RootPath, Mode: 0777}
	key := generator.String(20)
	if provider.Exists(key) {
		t.Error("Resource should not exist.")
	}
	provider.Write(key, s)
	if !provider.Exists(key) {
		t.Error("Resource should exist.")
	}
}

func TestDelete(t *testing.T) {
	target := &TestStruct{Foo: "foo", Bar: 20}
	provider := &disk{RootPath: RootPath, Mode: 0777}
	key := generator.String(20)
	provider.Write(key, target)
	provider.Delete(key)
	if provider.Exists(key) {
		t.Error("Resource should have been removed.")
	}
}

func Test_diskProvider_Shard(t *testing.T) {
	assert := assert.New(t)
	storage, err := NewDisk(RootPath, 0777)
	assert.NoError(err)
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
