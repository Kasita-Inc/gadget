package storage

import (
	"encoding/gob"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sync"

	"github.com/Kasita-Inc/gadget/crypto"
	"github.com/Kasita-Inc/gadget/errors"
	"github.com/Kasita-Inc/gadget/fileutil"
	"github.com/Kasita-Inc/gadget/stringutil"
)

const (
	// KeySuffix is the suffix appended to key names on disk.
	KeySuffix = ".gob"
)

// KeyFormat is used for creating the filename of the keys on disk
var KeyFormat = fmt.Sprintf("%%s%s", KeySuffix)

type disk struct {
	RootPath string
	Mode     os.FileMode
	mutex    sync.RWMutex
}

// NewDisk storage provider using the passed root path and creating files with the specified mode.
func NewDisk(rootPath string, mode os.FileMode) (Provider, error) {
	provider := &disk{RootPath: rootPath, Mode: mode}
	if !path.IsAbs(provider.RootPath) {
		return provider, fmt.Errorf("root path must be an absolute path, received '%s'", provider.RootPath)
	}
	_, err := fileutil.EnsureDir(provider.RootPath, provider.Mode)
	return provider, err
}

func (provider *disk) Shard(name string) (Provider, error) {
	return NewDisk(path.Join(provider.RootPath, name), provider.Mode)
}

func (provider *disk) getFullPath(key string) (string, error) {
	if stringutil.IsWhiteSpace(key) {
		return "", fmt.Errorf("invalid key (empty)")
	}
	filename := fmt.Sprintf(KeyFormat, keyToFilename(key))
	return path.Join(provider.RootPath, filename), nil
}

// Read into the target for the passed key.
func (provider *disk) Read(key string, target interface{}) error {
	// Put existence check in the persistence layer
	filepath, err := provider.getFullPath(key)
	if nil != err {
		return err
	}
	if !fileutil.FileExists(filepath) {
		return errors.New(NoEntryForKey)
	}
	provider.mutex.RLock()
	defer provider.mutex.RUnlock()
	var file *os.File
	if file, err = os.Open(filepath); nil != err {
		return err
	}
	defer file.Close()
	decoder := gob.NewDecoder(file)
	return decoder.Decode(target)
}

func keyToFilename(key string) string {
	// we don't need 512 byte filenames
	return crypto.Hash(key, "")[:32]
}

// Write the passed target for the specified key
func (provider *disk) Write(key string, target interface{}) error {
	provider.mutex.Lock()
	defer provider.mutex.Unlock()
	// get the full path
	filePath, err := provider.getFullPath(key)
	if nil != err {
		return err
	}

	fileutil.EnsureDir(filepath.Dir(filePath), provider.Mode)

	// create the file, this will overwrite if it already exists.
	w, err := os.Create(filePath)
	if nil != err {
		return err
	}
	defer w.Close()
	encoder := gob.NewEncoder(w)
	return encoder.Encode(target)
}

// Exists verifies that a record exists for the passed key.
func (provider *disk) Exists(key string) bool {
	provider.mutex.RLock()
	defer provider.mutex.RUnlock()
	filepath, _ := provider.getFullPath(key)
	return fileutil.FileExists(filepath)
}

// Delete removes the partition from disk
func (provider *disk) Delete(key string) error {
	provider.mutex.Lock()
	defer provider.mutex.Unlock()
	filepath, err := provider.getFullPath(key)
	if nil == err {
		err = os.Remove(filepath)
	}
	return err
}
