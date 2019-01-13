package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/function61/home-automation-hub/pkg/hapitypes"
	"github.com/hashicorp/hcl"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	confFilePath = "conf.hcl"
)

func readConfigurationFile() (*hapitypes.ConfigFile, error) {
	merged, free, err := readAllConfFilesMerged()
	if err != nil {
		return nil, err
	}
	defer free()

	return parseConfiguration(merged)
}

func parseConfiguration(hclContent io.Reader) (*hapitypes.ConfigFile, error) {
	// read all files concatenated (you can catenate HCL files) into a single blob
	hclContentBuffered, err := ioutil.ReadAll(hclContent)
	if err != nil {
		return nil, err
	}

	// read & parse HCL to generic struct
	var v interface{}
	errHcl := hcl.Unmarshal(hclContentBuffered, &v)
	if errHcl != nil {
		return nil, fmt.Errorf("unable to parse HCL: %s", errHcl)
	}

	// re-encode the generic struct to JSON, so we can unmarshal without
	// having both JSON and HCL struct tags

	asJson, errToJson := json.MarshalIndent(v, "", "  ")
	if errToJson != nil {
		return nil, errToJson
	}

	jsonDecoder := json.NewDecoder(bytes.NewBuffer(asJson))
	jsonDecoder.DisallowUnknownFields()

	conf := &hapitypes.ConfigFile{}
	if err := jsonDecoder.Decode(conf); err != nil {
		return nil, err
	}

	return conf, nil
}

func readAllConfFilesMerged() (io.Reader, func(), error) {
	noop := func() {}

	confFilePaths, err := filepath.Glob("conf/*.hcl")
	if err != nil {
		return nil, noop, err
	}

	confFiles := []*os.File{}
	readers := []io.Reader{}
	for _, confFilePath := range confFilePaths {
		confFile, err := os.Open(confFilePath)
		if err != nil {
			return nil, noop, err
		}

		confFiles = append(confFiles, confFile)
		readers = append(readers, confFile)
	}

	return io.MultiReader(readers...), func() {
		for _, r := range confFiles {
			r.Close()
		}
	}, nil
}

func findDeviceConfig(id string, conf *hapitypes.ConfigFile) *hapitypes.DeviceConfig {
	for _, devs := range conf.Devices {
		if devs.DeviceId == id {
			return &devs
		}
	}

	return nil
}
