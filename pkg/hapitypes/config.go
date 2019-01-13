package hapitypes

type PresenceByPingDevice struct {
	Ip     string `json:"ip"`
	Person string `json:"person"`
}

// always prefix your keys with <type> of your adapter.
// <type> should be same as pkg/adapters/<name>adapter (without the "adapter" suffix).

type AdapterConfig struct {
	Id   string `json:"id"`
	Type string `json:"type"`

	ParticleId          string `json:"particle_id,omitempty"`
	ParticleAccessToken string `json:"particle_access_token,omitempty"`

	HarmonyAddr string `json:"harmony_addr,omitempty"`

	SqsQueueUrl           string `json:"sqs_queue_url,omitempty"`
	SqsKeyId              string `json:"sqs_key_id,omitempty"`
	SqsKeySecret          string `json:"sqs_key_secret,omitempty"`
	SqsAlexaUsertokenHash string `json:"sqs_alexa_usertoken_hash,omitempty"`

	IrSimulatorKey string `json:"irsimulator_button,omitempty"`

	EventghostAddr   string `json:"eventghost_addr,omitempty"`
	EventghostSecret string `json:"eventghost_secret,omitempty"`

	TradfriUrl  string `json:"tradfri_url"`
	TradfriUser string `json:"tradfri_user"`
	TradfriPsk  string `json:"tradfri_psk"`

	Zigbee2MqttAddr string `json:"zigbee2mqtt_addr"`

	PresenceByPingDevice []PresenceByPingDevice `json:"presencebypingdevice"`

	DevicegroupDevices []string `json:"devicegroup_devs"`
}

type DeviceConfig struct {
	DeviceId         string `json:"id"`
	Type             string `json:"type"`
	AdapterId        string `json:"adapter"`
	AdaptersDeviceId string `json:"adapters_device_id,omitempty"`
	Name             string `json:"name"`
	Description      string `json:"description"`
	PowerOnCmd       string `json:"power_on_cmd,omitempty"`
	PowerOffCmd      string `json:"power_off_cmd,omitempty"`
	AlexaCategory    string `json:"alexa_category,omitempty"`
}

type Person struct {
	Id string `json:"id"`
}

type ActionConfig struct {
	Device    string `json:"device"`
	Verb      string `json:"verb"` // powerOn/powerOff/powerToggle/blink/ir
	IrCommand string `json:"ir_command"`
}

type SubscribeConfig struct {
	Event   string         `json:"event"`
	Actions []ActionConfig `json:"action"`
}

type ConfigFile struct {
	Adapters      []AdapterConfig   `json:"adapter"`
	Devices       []DeviceConfig    `json:"device"`
	Persons       []Person          `json:"person"`
	Subscriptions []SubscribeConfig `json:"subscribe"`
}

func (c *ConfigFile) FindDeviceConfigByAdaptersDeviceId(adaptersDeviceId string) *DeviceConfig {
	for _, deviceConfig := range c.Devices {
		if deviceConfig.AdaptersDeviceId == adaptersDeviceId {
			return &deviceConfig
		}
	}

	return nil
}
