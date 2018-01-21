package main

import (
	"encoding/json"
	"os"
)

const (
	confFilePath = "conf.json"
)

type AdapterConfig struct {
	Id   string `json:"id"`
	Type string `json:"type"`

	ParticleId          string `json:"particle_id,omitempty"`
	ParticleAccessToken string `json:"particle_access_token,omitempty"`

	HappyLightsAddr string `json:"happylights_addr,omitempty"`

	HarmonyAddr string `json:"harmony_addr,omitempty"`

	SqsQueueUrl  string `json:"sqs_queue_url,omitempty"`
	SqsKeyId     string `json:"sqs_key_id,omitempty"`
	SqsKeySecret string `json:"sqs_key_secret,omitempty"`

	IrSimulatorKey string `json:"irsimulator_button,omitempty"`
}

type DeviceConfig struct {
	DeviceId         string `json:"id"`
	AdapterId        string `json:"adapter"`
	AdaptersDeviceId string `json:"adapters_device_id,omitempty"`
	// used for Alexa
	Name        string `json:"name"`
	Description string `json:"description"`
	PowerOnCmd  string `json:"power_on_cmd,omitempty"`
	PowerOffCmd string `json:"power_off_cmd,omitempty"`
}

type DeviceGroupConfig struct {
	Id        string   `json:"id"`
	Name      string   `json:"name"`
	DeviceIds []string `json:"device_ids"`
}

type ConfigFile struct {
	Adapters     []AdapterConfig     `json:"adapter"`
	Devices      []DeviceConfig      `json:"device"`
	DeviceGroups []DeviceGroupConfig `json:"devicegroup"`
}

func writeConfigurationFile(conf *ConfigFile) error {
	file, err := os.Create(confFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	if err := enc.Encode(&conf); err != nil {
		return err
	}

	return nil
}

func readConfigurationFile() (*ConfigFile, error) {
	file, err := os.Open(confFilePath)
	if err != nil {
		return nil, err
	}

	var conf ConfigFile
	dec := json.NewDecoder(file)
	if err := dec.Decode(&conf); err != nil {
		return nil, err
	}

	return &conf, nil
}
