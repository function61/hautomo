package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/function61/home-automation-hub/hapitypes"
	"github.com/hashicorp/hcl"
	"io/ioutil"
)

const (
	confFilePath = "conf.hcl"
)

func readConfigurationFile() (*hapitypes.ConfigFile, error) {
	// read & parse HCL to generic struct
	input, err := ioutil.ReadFile(confFilePath)
	if err != nil {
		return nil, err
	}

	var v interface{}
	errHcl := hcl.Unmarshal(input, &v)
	if errHcl != nil {
		return nil, fmt.Errorf("unable to parse HCL: %s", errHcl)
	}

	// re-encode the generic struct to JSON, so we can unmarshal without
	// having both JSON and HCL struct tags

	asJson, errToJson := json.MarshalIndent(v, "", "  ")
	if errToJson != nil {
		return nil, errToJson
	}

	// decode JSON to in-memory config struct

	var conf hapitypes.ConfigFile
	dec := json.NewDecoder(bytes.NewBuffer(asJson))
	if err := dec.Decode(&conf); err != nil {
		return nil, err
	}

	return &conf, nil
}
