package utilpkg

import (
	"errors"
	"io/ioutil"
	"os"

	"github.com/go-yaml/yaml"
)

// GetConfig takes <path string> <output interface{}>
// Reads configuration
// Returns error
func GetConfig(path string, output interface{}) error {
	if _, err := os.Stat(path); err != nil {
		return genConfig(path, output)
	}

	// Config file exists, so we're reading it.
	file, err := ioutil.ReadFile(path)

	if err != nil {
		return err
	}

	// Parse the yml file
	_ = yaml.Unmarshal(file, output)

	return nil
}

// genConfig takes <path string> <reference interface{}>
// Generates configuration
// Returns error
func genConfig(path string, reference interface{}) error {
	serialized, err := yaml.Marshal(reference)

	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path, serialized, 0660)

	if err != nil {
		return err
	}

	return errors.New("generated a new default configuration")
}