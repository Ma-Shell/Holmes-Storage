package ObjStorerLocalFS

import (
	"io/ioutil"
	"errors"
	"os"
	"path/filepath"
	"bytes"
	"strings"

	"github.com/HolmesProcessing/Holmes-Storage/objStorerGeneric"
)

type ObjStorerLocalFS struct {
	StorageLocation string
	ConfigStorageLocation string
}

func (s ObjStorerLocalFS) Initialize(configs []*objStorerGeneric.ObjDBConnector) (objStorerGeneric.ObjStorer, error) {
	// check for storage location setting
	// re-using the Bucket setting, local-fs is not the preferred storage method
	// and as such we don't encourage it by adding a dedicated setting
	if len(configs) > 0 && configs[0].Bucket != "" {
		s.StorageLocation = configs[0].Bucket
	} else {
		s.StorageLocation = "./objstorage-local-fs"
	}
	var configStorageLocation string
	if len(configs) > 0 && configs[0].ConfigBucket != "" {
		configStorageLocation = configs[0].ConfigBucket
	} else {
		configStorageLocation = "./configstorage-local-fs"
	}

	configStorageLocation = filepath.Clean(configStorageLocation)
	s.ConfigStorageLocation = configStorageLocation
	
	// setup storage location if not exists
	if err := s.Setup(); err != nil {
		return s, err
	}
	
	// create a temporary file to test writing + reading
	data := []byte("test content")
	path := filepath.Join(s.StorageLocation, "tempfile")
	
	// test writing
	if err := ioutil.WriteFile(path, data, 0644); err != nil {
		os.Remove(path)
		return s, err
	}
	
	// test reading
	if data2, err := ioutil.ReadFile(path); err != nil {
		os.Remove(path)
		return s, err
	} else if !bytes.Equal(data, data2) {
		os.Remove(path)
		return s, errors.New("tempfile write/read failed, data mismatch")
	}
	
	// test removal
	if err := os.Remove(path); err != nil {
		return s, err
	}
	
	return s, nil
}

func (s ObjStorerLocalFS) Setup() error {
	err := os.MkdirAll(s.StorageLocation, 0755)
	if err != nil {
		return err
	}
	_, err = os.Stat(s.StorageLocation)
	if err != nil {
		return err
	}

	err = os.MkdirAll(s.ConfigStorageLocation, 0755)
	if err != nil {
		return err
	}
	_, err = os.Stat(s.ConfigStorageLocation)
	return err
}

func (s ObjStorerLocalFS) StoreSample(sample *objStorerGeneric.Sample) error {
	path := filepath.Join(s.StorageLocation, sample.SHA256)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return ioutil.WriteFile(path, sample.Data, 0644)
	} else if os.IsPermission(err) {
		return errors.New("permission denied")
	} else if os.IsExist(err) {
		return nil  // duplicates are fine
	} else {
		return err
	}
}

func (s ObjStorerLocalFS) GetSample(id string) (*objStorerGeneric.Sample, error) {
	sample := &objStorerGeneric.Sample{SHA256: id}
	path := filepath.Join(s.StorageLocation, sample.SHA256)
	data, err := ioutil.ReadFile(path)
	sample.Data = data
	return sample, err
}

// TODO: Support MultipleObjects retrieval and getting. Useful when using something over 100megs

func (s ObjStorerLocalFS) SanitizePath(path string) (string, error) {
	path = filepath.Join(s.ConfigStorageLocation, path)
	if !strings.HasPrefix(path, s.ConfigStorageLocation) {
		return "", errors.New("permission denied")
	}
	return path, nil
}

func (s ObjStorerLocalFS) StoreConfig(config * objStorerGeneric.Config) error {
	path, err := s.SanitizePath(config.Path)

	// Make sure all the parent directories exist
	err = os.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		return err
	}

	// If the file already exists, simply overwrite it
	return ioutil.WriteFile(path, config.FileContents, 0644)
}

func (s ObjStorerLocalFS) GetConfig(path string) (*objStorerGeneric.Config, error) {
	config := &objStorerGeneric.Config{Path: path}
	path, err := s.SanitizePath(path)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadFile(path)
	config.FileContents = data
	return config, err
}
