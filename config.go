package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/hcl"
	"io/ioutil"
)

const (
	confFilePath = "conf.hcl"
)

type AdapterConfig struct {
	Id   string `json:"id"`
	Type string `json:"type"`

	ParticleId          string `json:"particle_id,omitempty"`
	ParticleAccessToken string `json:"particle_access_token,omitempty"`

	HappyLightsAddr string `json:"happylights_addr,omitempty"`

	HarmonyAddr string `json:"harmony_addr,omitempty"`

	SqsQueueUrl           string `json:"sqs_queue_url,omitempty"`
	SqsKeyId              string `json:"sqs_key_id,omitempty"`
	SqsKeySecret          string `json:"sqs_key_secret,omitempty"`
	SqsAlexaUsertokenHash string `json:"sqs_alexa_usertoken_hash,omitempty"`

	IrSimulatorKey string `json:"irsimulator_button,omitempty"`

	EventghostAddr   string `json:"eventghost_addr,omitempty"`
	EventghostSecret string `json:"eventghost_secret,omitempty"`
}

type DeviceConfig struct {
	DeviceId         string `json:"id"`
	AdapterId        string `json:"adapter"`
	AdaptersDeviceId string `json:"adapters_device_id,omitempty"`
	// used for Alexa
	Name              string   `json:"name"`
	Description       string   `json:"description"`
	PowerOnCmd        string   `json:"power_on_cmd,omitempty"`
	PowerOffCmd       string   `json:"power_off_cmd,omitempty"`
	AlexaCategory     string   `json:"alexa_category,omitempty"`
	AlexaCapabilities []string `json:"alexa_capabilities,omitempty"`
}

type DeviceGroupConfig struct {
	Id        string   `json:"id"`
	Name      string   `json:"name"`
	DeviceIds []string `json:"device_ids"`
}

type IrPowerConfig struct {
	RemoteKey string `json:"remote_key"`
	ToDevice  string `json:"to_device"`
	PowerKind string `json:"power_kind"`
}

type IrToIr struct {
	RemoteKey string `json:"remote_key"`
	ToDevice  string `json:"to_device"`
	IrEvent   string `json:"ir"`
}

type ConfigFile struct {
	Adapters     []AdapterConfig     `json:"adapter"`
	Devices      []DeviceConfig      `json:"device"`
	DeviceGroups []DeviceGroupConfig `json:"devicegroup"`
	IrPowers     []IrPowerConfig     `json:"ir_powers"`
	IrToIr       []IrToIr            `json:"ir2ir"`
}

func readConfigurationFile() (*ConfigFile, error) {
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

	var conf ConfigFile
	dec := json.NewDecoder(bytes.NewBuffer(asJson))
	if err := dec.Decode(&conf); err != nil {
		return nil, err
	}

	return &conf, nil
}
