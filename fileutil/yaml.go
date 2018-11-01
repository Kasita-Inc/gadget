package fileutil

import (
	"io"
	"io/ioutil"

	"gopkg.in/yaml.v2"

	"github.com/Kasita-Inc/gadget/buffer"
)

// ReadYamlFromFile at path filename into the target interface.
func ReadYamlFromFile(filename string, target interface{}) error {
	data, err := ioutil.ReadFile(filename)
	if nil != err {
		return err
	}
	return yaml.Unmarshal(data, target)
}

// WriteYamlToFile at path filename sourcing the data from the passed target.
func WriteYamlToFile(filename string, target interface{}) error {
	data, err := yaml.Marshal(target)
	if nil != err {
		return err
	}
	return ioutil.WriteFile(filename, data, 0777)
}

// WriteYamlToWriter returning any errors that occur.
func WriteYamlToWriter(writer io.Writer, target interface{}) error {
	data, err := yaml.Marshal(target)
	if nil != err {
		return err
	}
	return buffer.Write(writer, data)
}
