package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"

	"github.com/fiamma-chain/fiamma-go-sdk/errors"
)

// LoadYAML config into out interface, with defaults and validates
func LoadYAML(path string, out interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return errors.Trace(err)
	}
	res, err := ParseEnv(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "config parse error: %s", err.Error())
		res = data
	}
	return UnmarshalYAML(res, out)
}

// ParseEnv parses env
func ParseEnv(data []byte) ([]byte, error) {
	text := string(data)
	envs := os.Environ()
	envMap := make(map[string]string)
	for _, s := range envs {
		t := strings.Split(s, "=")
		envMap[t[0]] = t[1]
	}
	tmpl, err := template.New("template").Option("missingkey=error").Parse(text)
	if err != nil {
		return nil, errors.Trace(err)
	}
	buffer := bytes.NewBuffer(nil)
	err = tmpl.Execute(buffer, envMap)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return buffer.Bytes(), nil
}

// UnmarshalYAML unmarshals, defaults and validates
func UnmarshalYAML(in []byte, out interface{}) error {
	err := yaml.Unmarshal(in, out)
	if err != nil {
		return errors.Trace(err)
	}
	err = SetDefaults(out)
	if err != nil {
		return errors.Trace(err)
	}
	return nil
}

// UnmarshalJSON unmarshals, defaults and validates
func UnmarshalJSON(in []byte, out interface{}) error {
	err := json.Unmarshal(in, out)
	if err != nil {
		return errors.Trace(err)
	}
	err = SetDefaults(out)
	if err != nil {
		return errors.Trace(err)
	}
	return nil
}
