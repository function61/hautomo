package main

import (
	"io"
	"os"
	"path/filepath"

	"github.com/function61/gokit/hcl2json"
	"github.com/function61/gokit/jsonfile"
	"github.com/function61/hautomo/pkg/hapitypes"
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
	// transform HCL into JSON, so we can unmarshal without having both JSON and HCL struct tags
	hclAsJson, err := hcl2json.Convert(hclContent)
	if err != nil {
		return nil, err
	}

	conf := &hapitypes.ConfigFile{}
	return conf, jsonfile.Unmarshal(hclAsJson, conf, true)
}

func readAllConfFilesMerged() (io.Reader, func(), error) {
	noop := func() {}

	// read all files concatenated (you can catenate HCL files) into a single blob
	confFilePaths, err := filepath.Glob("conf/*.hcl")
	if err != nil {
		return nil, noop, err
	}

	// two slices of same thing because we cannot typecast slice to another
	// https://stackoverflow.com/a/12754757
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
