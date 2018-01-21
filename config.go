package main

type AdapterConfig struct {
	Id   string `json:"Id"`
	Type string `json:"Type"`

	ParticleId          string `json:"ParticleId"`
	ParticleAccessToken string `json:"ParticleAccessToken"`

	HarmonyAddr string `json:"HarmonyAddr"`

	SqsQueueUrl  string `json:"SqsQueueUrl"`
	SqsKeyId     string `json:"SqsKeyId"`
	SqsKeySecret string `json:"SqsKeySecret"`

	IrSimulatorKey string `json:"IrSimulatorKey"`
}

type DeviceConfig struct {
	DeviceId         string `json:"DeviceId"`
	AdapterId        string `json:"AdapterId"`
	AdaptersDeviceId string `json:"AdaptersDeviceId"`
	// used for Alexa
	Name        string
	Description string `json:"Description"`
	PowerOnCmd  string `json:"PowerOnCmd"`
	PowerOffCmd string `json:"PowerOffCmd"`
}

type DeviceGroupConfig struct {
	Id        string   `json:"Id"`
	Name      string   `json:"Name"`
	DeviceIds []string `json:"DeviceIds"`
}

type ConfigFile struct {
	Adapters     []AdapterConfig     `json:"Adapters"`
	Devices      []DeviceConfig      `json:"Devices"`
	DeviceGroups []DeviceGroupConfig `json:"DeviceGroups"`
}
