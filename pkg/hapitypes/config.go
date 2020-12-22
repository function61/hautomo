package hapitypes

type PresenceByPingDevice struct {
	Ip     string `json:"ip"`
	Person string `json:"person"`
}

// always prefix your keys with <type> of your adapter.
// <type> should be same as pkg/adapters/<name>adapter/ (without the "adapter" suffix).

type AdapterConfig struct {
	Id   string `json:"id"`
	Type string `json:"type"`

	// generic attributes usable by many adapters
	Url string `json:"url"` // (or base URL), used by: alexa | screenserver | homeassistant | tradfi | zigbee2mqtt | harmonyhub

	ParticleId          string `json:"particle_id,omitempty"`
	ParticleAccessToken string `json:"particle_access_token,omitempty"`

	SqsKeyId                string `json:"sqs_key_id,omitempty"`
	SqsKeySecret            string `json:"sqs_key_secret,omitempty"`
	SqsAlexaUsertokenHash   string `json:"sqs_alexa_usertoken_hash,omitempty"`
	AlexaOauth2ClientId     string `json:"alexa_oauth2_client_id"`
	AlexaOauth2ClientSecret string `json:"alexa_oauth2_client_secret"`
	AlexaOauth2UserToken    string `json:"alexa_oauth2_user_token"`

	IrSimulatorKey string `json:"irsimulator_button,omitempty"`

	TradfriUser string `json:"tradfri_user"`
	TradfriPsk  string `json:"tradfri_psk"`

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

	EventghostAddr   string `json:"eventghost_addr,omitempty"` // if specified, we connect to the PC direction for sending events
	EventghostSecret string `json:"eventghost_secret,omitempty"`
}

// these are transparently generated to adapter + device combo
type DeviceGroupConfig struct {
	DeviceId string   `json:"device_id"`
	Name     string   `json:"name"`
	Devices  []string `json:"devices"`
}

type Person struct {
	Id string `json:"id"`
}

type ActionConfig struct {
	Device          string `json:"device"`
	Verb            string `json:"verb"`             // powerOn/powerOff/powerToggle/blink/ir/setBooleanFalse/setBooleanTrue/sleep/playback/notify/speak
	IrCommand       string `json:"ir_command"`       // used by: ir
	Boolean         string `json:"boolean"`          // used by: setBooleanTrue/setBooleanFalse
	DurationSeconds int    `json:"duration_seconds"` // used by: sleep
	PlaybackAction  string `json:"playback_action"`  // used by: playback
	NotifyMessage   string `json:"notify_message"`   // used by: notify
	SpeakPhrase     string `json:"speak_phrase"`     // used by: speak
}

type ConditionConfig struct {
	Type            string `json:"type"` // boolean-is-true/boolean-is-false/boolean-not-changed-within
	Boolean         string `json:"boolean"`
	DurationSeconds int    `json:"duration_seconds"`
}

type SubscribeConfig struct {
	Event      string            `json:"event"`
	Actions    []ActionConfig    `json:"action"`
	Conditions []ConditionConfig `json:"condition"`
}

type ConfigFile struct {
	Adapters      []AdapterConfig     `json:"adapter"`
	Devices       []DeviceConfig      `json:"device"`
	DeviceGroups  []DeviceGroupConfig `json:"devicegroup"`
	Persons       []Person            `json:"person"`
	Subscriptions []SubscribeConfig   `json:"subscribe"`
}

func (c *ConfigFile) FindDeviceConfigByAdaptersDeviceId(adaptersDeviceId string) *DeviceConfig {
	for _, deviceConfig := range c.Devices {
		if deviceConfig.AdaptersDeviceId == adaptersDeviceId {
			return &deviceConfig
		}
	}

	return nil
}
