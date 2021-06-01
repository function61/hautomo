package cluster

// Local command is cluster-specific command and its arguments.
//
// E.G. CommandClusterAndId=(cluster=On/Off command=Off With Effect) and its arguments
// are struct fields EffectIdentifier and EffectVariant
type LocalCommand interface {
	CommandClusterAndId() (ClusterId, uint8)
}

// These struct fields, their order and their sizes are defined by ZCL spec.
// E.G. for *TriggerEffectCommand* see ZCL spec section 3.5.2.3.3
//
// pro-tip: you can search below-mentioned ZCL spec section numbers from the spec:
//   https://zigbeealliance.org/wp-content/uploads/2019/12/07-5123-06-zigbee-cluster-library-specification.pdf

type ResetToFactoryDefaultsCommand struct {
}

// -------- Cluster: Identify --------

type IdentifyCommand struct {
	IdentifyTime uint16
}

type IdentifyQueryCommand struct{}

type EffectId uint8

const (
	EffectIdBlink         EffectId = 0x00 // Light is turned on/off once
	EffectIdBreathe       EffectId = 0x01 // Light turned on/off over 1 second and repeated 15 times.
	EffectIdOkay          EffectId = 0x02 // Colored light turns green for 1 second; noncolored light flashes twice
	EffectIdChannelChange EffectId = 0x0b // Colored light turns orange for 8 seconds; noncolored light switches to maximum brightness for 0.5s and then minimum brightness for 7.5s.
	EffectIdFinishEffect  EffectId = 0xfe // Complete the current effect sequence before terminating. e.g., if in the middle of a breathe effect (as above), first complete the current 1s breathe effect and then terminate the effect
	EffectIdStopEffect    EffectId = 0xff // Terminate the effect as soon as possible
)

// ZCL spec section: 3.5.2.3.3
type GenIdentifyTriggerEffectCommand struct {
	Effect        EffectId
	EffectVariant uint8 // variant of effect *EffectIdentifier*. usually zero
}

func (c *GenIdentifyTriggerEffectCommand) CommandClusterAndId() (ClusterId, uint8) {
	return IdGenIdentify, 0x40
}

var _ LocalCommand = (*GenIdentifyTriggerEffectCommand)(nil)

type IdentifyQueryResponse struct {
	Timeout uint16
}

// -------- Cluster: GenOnOff --------

// ZCL spec section: 3.8.2.3.1
type GenOnOffOffCommand struct{}

func (c *GenOnOffOffCommand) CommandClusterAndId() (ClusterId, uint8) {
	return IdGenOnOff, 0x00
}

var _ LocalCommand = (*GenOnOffOffCommand)(nil)

// ZCL spec section: 3.8.2.3.2
type GenOnOffOnCommand struct{}

func (c *GenOnOffOnCommand) CommandClusterAndId() (ClusterId, uint8) {
	return IdGenOnOff, 0x01
}

var _ LocalCommand = (*GenOnOffOnCommand)(nil)

// ZCL spec section: 3.8.2.3.3
type GenOnOffToggleCommand struct{}

func (c *GenOnOffToggleCommand) CommandClusterAndId() (ClusterId, uint8) {
	return IdGenOnOff, 0x02
}

var _ LocalCommand = (*GenOnOffToggleCommand)(nil)

type OffWithEffectCommand struct {
	EffectIdentifier uint8
	EffectVariant    uint8
}

type OnWithRecallGlobalSceneCommand struct{}

type OnWithTimedOffCommand struct {
	OnOffControl uint8
	OnTime       uint16
	OffWaitTime  uint16
}

// -------- Cluster: IdGenLevelCtrl --------

// ZCL spec section: 3.10.2.4.1
type MoveToLevelCommand struct {
	Level          uint8  // valid range (0x00, 0xfe)
	TransitionTime uint16 // [100ms], valid range (0x0000, 0xffff)
}

func (c *MoveToLevelCommand) CommandClusterAndId() (ClusterId, uint8) {
	return IdGenLevelCtrl, 0x00
}

var _ LocalCommand = (*MoveToLevelCommand)(nil)

// ZCL spec section: 3.10.2.4.2
type MoveCommand struct {
	MoveMode uint8
	Rate     uint8
}

func (c *MoveCommand) CommandClusterAndId() (ClusterId, uint8) {
	return IdGenLevelCtrl, 0x01
}

var _ LocalCommand = (*MoveCommand)(nil)

// ZCL spec section: 3.10.2.4.3
type StepCommand struct {
	StepMode       uint8
	StepSize       uint8
	TransitionTime uint16
}

func (c *StepCommand) CommandClusterAndId() (ClusterId, uint8) {
	return IdGenLevelCtrl, 0x02
}

var _ LocalCommand = (*StepCommand)(nil)

// ZCL spec section: 3.10.2.4.4
type StopCommand struct{}

func (c *StopCommand) CommandClusterAndId() (ClusterId, uint8) {
	return IdGenLevelCtrl, 0x03
}

var _ LocalCommand = (*StopCommand)(nil)

// ZCL spec section: 3.10.2.4.5
type MoveToLevelOnOffCommand struct {
	Level          uint8
	TransitionTime uint16
}

func (c *MoveToLevelOnOffCommand) CommandClusterAndId() (ClusterId, uint8) {
	return IdGenLevelCtrl, 0x04
}

var _ LocalCommand = (*MoveToLevelOnOffCommand)(nil)

// ZCL spec section: 3.10.2.4.5
type MoveOnOffCommand struct {
	MoveMode uint8
	Rate     uint8
}

func (c *MoveOnOffCommand) CommandClusterAndId() (ClusterId, uint8) {
	return IdGenLevelCtrl, 0x05
}

var _ LocalCommand = (*MoveOnOffCommand)(nil)

// ZCL spec section: 3.10.2.4.5
type StepOnOffCommand struct {
	StepMode       uint8
	StepSize       uint8
	TransitionTime uint16
}

func (c *StepOnOffCommand) CommandClusterAndId() (ClusterId, uint8) {
	return IdGenLevelCtrl, 0x06
}

var _ LocalCommand = (*StepOnOffCommand)(nil)

// 0x07,StopOnOff, ZCL spec section: 3.10.2.4.5
type StopOnOffCommand struct{}

func (c *StopOnOffCommand) CommandClusterAndId() (ClusterId, uint8) {
	return IdGenLevelCtrl, 0x07
}

var _ LocalCommand = (*StopOnOffCommand)(nil)

// -------- Cluster: LightingColorCtrl --------

// ZCL spec section: 5.2.2.3.10
type LightingColorCtrlMoveToColor struct {
	X              uint16
	Y              uint16
	TransitionTime uint16
}

func (c *LightingColorCtrlMoveToColor) CommandClusterAndId() (ClusterId, uint8) {
	return IdLightingColorCtrl, 0x07
}

var _ LocalCommand = (*LightingColorCtrlMoveToColor)(nil)

// ZCL spec section: 5.2.2.3.13
type LightingColorCtrlMoveToColorTemperature struct {
	ColorTemperatureMireds uint16 // Blue sky = 40 mireds, photography flash = 200 mireds. Candle flame = 667. http://www.photokaboom.com/photography/learn/Photoshop_Elements/color/1_color_temperature_mired.htm valid range (0x0000, 0xfeff)
	TransitionTime         uint16
}

func (c *LightingColorCtrlMoveToColorTemperature) CommandClusterAndId() (ClusterId, uint8) {
	return IdLightingColorCtrl, 0x0a
}

var _ LocalCommand = (*LightingColorCtrlMoveToColorTemperature)(nil)

// -------- Cluster: IdClosuresWindowCovering --------

// ZCL spec section: 7.4.2.2.1
type ClosuresWindowCoveringUp struct{}

func (c *ClosuresWindowCoveringUp) CommandClusterAndId() (ClusterId, uint8) {
	return IdClosuresWindowCovering, 0x00
}

var _ LocalCommand = (*ClosuresWindowCoveringUp)(nil)

// ZCL spec section: 7.4.2.2.2
type ClosuresWindowCoveringDown struct{}

func (c *ClosuresWindowCoveringDown) CommandClusterAndId() (ClusterId, uint8) {
	return IdClosuresWindowCovering, 0x01
}

var _ LocalCommand = (*ClosuresWindowCoveringDown)(nil)

// ZCL spec section: 7.4.2.2.3
type ClosuresWindowCoveringStop struct{}

func (c *ClosuresWindowCoveringStop) CommandClusterAndId() (ClusterId, uint8) {
	return IdClosuresWindowCovering, 0x02
}

var _ LocalCommand = (*ClosuresWindowCoveringStop)(nil)

// ZCL spec section: 7.4.2.2.4
type ClosuresWindowCoveringGoToLiftPercentage struct {
	Value uint8 // 0-100
}

func (c *ClosuresWindowCoveringGoToLiftPercentage) CommandClusterAndId() (ClusterId, uint8) {
	return IdClosuresWindowCovering, 0x05
}

var _ LocalCommand = (*ClosuresWindowCoveringGoToLiftPercentage)(nil)

// -------- Cluster: Scenes --------

// undocumented in ZCL spec...........
type ScenesMysteryCommand7 struct {
	DataRaw [4]byte
}

func (c *ScenesMysteryCommand7) CommandClusterAndId() (ClusterId, uint8) {
	return IdGenOnOff, 0x07
}

func (c *ScenesMysteryCommand7) Left() bool {
	// first byte 0x01 if going to "left" scene, 0x00 if "right"
	return c.DataRaw[0] == 0x01
}

var _ LocalCommand = (*ScenesMysteryCommand7)(nil)

// the following are "generated" commands (end device -> coordinator direction)

type ZoneStatusChangeNotificationCommand struct {
	ZoneStatus     ZoneStatus
	ExtendedStatus uint8 // "SHALL be set to zero"
	ZoneID         uint8
	Delay          uint16
}

// ZCL spec section: 8.2.2.2.1.3
type ZoneStatus uint16

func (z ZoneStatus) Alarm1() bool {
	return z&0b1 != 0
}

func (z ZoneStatus) UnimplementedBitsSet() bool {
	return z|0b1 != 1
}
