package harmonyhub

import (
	"../../util/stopper"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"golang.org/x/net/html/charset"
	"log"
	"net"
	"time"
)

var errNotConnected = errors.New("not connected")

// https://github.com/petele/PiHomeControl
// http://petelepage.com/blog/2013/07/home-automation-for-geeks/

/*
	Client:
		<stream:stream xmlns="jabber:client" xmlns:stream="http://etherx.jabber.org/streams" version="1.0" to="x.com">
	Server:
		<?xml version='1.0' encoding='iso-8859-1'?><stream:stream from='x.com' id='001418d7' version='1.0' xmlns='jabber:client' xmlns:stream='http://etherx.jabber.org/streams'><stream:features><mechanisms xmlns='urn:ietf:params:xml:ns:xmpp-sasl'><mechanism>PLAIN</mechanism></mechanisms></stream:features>

	Client:
		<auth xmlns="urn:ietf:params:xml:ns:xmpp-sasl" mechanism="PLAIN">Z3Vlc3RAeC5jb20AZ3Vlc3QAZ3Vlc3Q=</auth>
	Server:
		<success xmlns='urn:ietf:params:xml:ns:xmpp-sasl'/>

	Client:
		<stream:stream xmlns="jabber:client" xmlns:stream="http://etherx.jabber.org/streams" version="1.0" to="x.com">
	Server:
		<stream:stream from='x.com' id='001418d7' version='1.0' xmlns='jabber:client' xmlns:stream='http://etherx.jabber.org/streams'><stream:features><bind xmlns='urn:ietf:params:xml:ns:xmpp-bind'/><session xmlns='urn:ietf:params:xml:nx:xmpp-session'/></stream:features>

	Client:
		<iq type="set" id="bind"><bind xmlns="urn:ietf:params:xml:ns:xmpp-bind"><resource>gatorade</resource></bind></iq>
	Server:
		<iq id='bind' type='result'><bind xmlns='urn:ietf:params:xml:ns:xmpp-bind'><jid>1111/gatorade</jid></bind></iq>

	Client:
		<iq type="get" id="662413" from="guest"><oa xmlns="connect.logitech.com" mime="vnd.logitech.connect/vnd.logitech.pair">method=pair:name=harmonyjs#iOS6.0.1#iPhone</oa></iq>
	Server:
		<iq/><iq id="662413" to="guest" type="get"><oa xmlns='connect.logitech.com' mime='vnd.logitech.connect/vnd.logitech.pair' errorcode='200' errorstring='OK'><![CDATA[serverIdentity=95f8d8c4-8fe1-4712-bf49-a374ce916c45:hubId=106:identity=95f8d8c4-8fe1-4712-bf49-a374ce916c45:status=succeeded:protocolVersion={XMPP="1.0", HTTP="1.0", RF="1.0", WEBSOCKET="1.0"}:hubProfiles={Harmony="2.0"}:productId=Pimento:friendlyName=Harmony Hub]]></oa></iq>

	Client:
		</stream:stream>
	Server:
		</stream:stream>



TODO: CLI/API for getting config

<iq type="get" id="182632"><oa xmlns="connect.logitech.com" mime="vnd.logitech.harmony/vnd.logitech.harmony.engine?config"></oa></iq><iq/><iq id="182632" type="get"><oa xmlns='connect.logitech.com' mime='vnd.logitech.harmony/vnd.logitech.harmony.engine?config' errorcode='200' errorstring='OK'><![CDATA[{"activity":[{"rules":[],"label":"Watch TV","isTuningDefault":false,"activityTypeDisplayName":"Default","enterActions":[],"fixit":{"47917687":{"id":"47917687","Input":"Game","Power":"On"},"47918441":{"id":"47918441","Power":"On"}},"baseImageUri":"https:\/\/rcbu-test-ssl-amr.s3.amazonaws.com\/","zones":null,"type":"VirtualTelevisionN","isAVActivity":true,"id":"28648157","suggestedDisplay":"Default","controlGroup":[{"name":"NumericBasic","function":[{"action":"{\"command\":\"0\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"Number0","label":"0"},{"action":"{\"command\":\"1\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"Number1","label":"1"},{"action":"{\"command\":\"2\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"Number2","label":"2"},{"action":"{\"command\":\"3\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"Number3","label":"3"},{"action":"{\"command\":\"4\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"Number4","label":"4"},{"action":"{\"command\":\"5\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"Number5","label":"5"},{"action":"{\"command\":\"6\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"Number6","label":"6"},{"action":"{\"command\":\"7\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"Number7","label":"7"},{"action":"{\"command\":\"8\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"Number8","label":"8"},{"action":"{\"command\":\"9\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"Number9","label":"9"}]},{"name":"Volume","function":[{"action":"{\"command\":\"Mute\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"Mute","label":"Mute"},{"action":"{\"command\":\"VolumeDown\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"VolumeDown","label":"Volume Down"},{"action":"{\"command\":\"VolumeUp\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"VolumeUp","label":"Volume Up"}]},{"name":"Channel","function":[{"action":"{\"command\":\"ChannelPrev\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"PrevChannel","label":"Prev Channel"},{"action":"{\"command\":\"ChannelDown\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"ChannelDown","label":"Channel Down"},{"action":"{\"command\":\"ChannelUp\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"ChannelUp","label":"Channel Up"}]},{"name":"NavigationBasic","function":[{"action":"{\"command\":\"DirectionDown\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"DirectionDown","label":"Direction Down"},{"action":"{\"command\":\"DirectionLeft\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"DirectionLeft","label":"Direction Left"},{"action":"{\"command\":\"DirectionRight\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"DirectionRight","label":"Direction Right"},{"action":"{\"command\":\"DirectionUp\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"DirectionUp","label":"Direction Up"},{"action":"{\"command\":\"Ok\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"Select","label":"Select"}]},{"name":"TransportBasic","function":[{"action":"{\"command\":\"Stop\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"Stop","label":"Stop"},{"action":"{\"command\":\"Play\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"Play","label":"Play"},{"action":"{\"command\":\"Rewind\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"Rewind","label":"Rewind"},{"action":"{\"command\":\"Pause\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"Pause","label":"Pause"},{"action":"{\"command\":\"FastForward\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"FastForward","label":"Fast Forward"}]},{"name":"TransportRecording","function":[{"action":"{\"command\":\"Record\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"Record","label":"Record"}]},{"name":"NavigationDVD","function":[{"action":"{\"command\":\"Subtitle\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"Subtitle","label":"Subtitle"},{"action":"{\"command\":\"Back\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"Back","label":"Back"}]},{"name":"NavigationDSTB","function":[{"action":"{\"command\":\"List\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"List","label":"List"},{"action":"{\"command\":\"Search\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"Search","label":"Search"}]},{"name":"GameType3","function":[{"action":"{\"command\":\"Home\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"Home","label":"Home"}]},{"name":"NavigationExtended","function":[{"action":"{\"command\":\"Guide\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"Guide","label":"Guide"},{"action":"{\"command\":\"Info\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"Info","label":"Info"},{"action":"{\"command\":\"Exit\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"Exit","label":"Exit"},{"action":"{\"command\":\"PageDown\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"PageDown","label":"Page Down"},{"action":"{\"command\":\"PageUp\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"PageUp","label":"Page Up"}]},{"name":"DisplayMode","function":[{"action":"{\"command\":\"Aspect\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"Aspect","label":"Aspect"}]},{"name":"ColoredButtons","function":[{"action":"{\"command\":\"Green\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"Green","label":"Green"},{"action":"{\"command\":\"Red\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"Red","label":"Red"},{"action":"{\"command\":\"Blue\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"Blue","label":"Blue"},{"action":"{\"command\":\"Yellow\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"Yellow","label":"Yellow"}]},{"name":"Teletext","function":[{"action":"{\"command\":\"Teletext\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"Teletext","label":"Teletext"}]}],"sequences":[],"activityOrder":1,"roles":{"ChannelChangingActivityRole":"47918441","DisplayActivityRole":"47918441","VolumeActivityRole":"47917687"},"VolumeActivityRole":"47917687","isMultiZone":false,"icon":"userdata: 0x4454e0"},{"rules":[],"label":"Test","isTuningDefault":false,"activityTypeDisplayName":"Default","enterActions":[],"fixit":{"47917687":{"id":"47917687","Input":"Game","Power":"On"},"47918441":{"id":"47918441","Power":"Off"}},"icon":"userdata: 0x4454e0","zones":null,"suggestedDisplay":"Default","isAVActivity":true,"id":"28647596","type":"VirtualCdMulti","sequences":[],"controlGroup":[{"name":"NumericBasic","function":[{"action":"{\"command\":\"0\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"Number0","label":"0"},{"action":"{\"command\":\"1\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"Number1","label":"1"},{"action":"{\"command\":\"2\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"Number2","label":"2"},{"action":"{\"command\":\"3\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"Number3","label":"3"},{"action":"{\"command\":\"4\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"Number4","label":"4"},{"action":"{\"command\":\"5\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"Number5","label":"5"},{"action":"{\"command\":\"6\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"Number6","label":"6"},{"action":"{\"command\":\"7\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"Number7","label":"7"},{"action":"{\"command\":\"8\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"Number8","label":"8"},{"action":"{\"command\":\"9\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"Number9","label":"9"}]},{"name":"Volume","function":[{"action":"{\"command\":\"Mute\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"Mute","label":"Mute"},{"action":"{\"command\":\"VolumeDown\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"VolumeDown","label":"Volume Down"},{"action":"{\"command\":\"VolumeUp\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"VolumeUp","label":"Volume Up"}]},{"name":"NavigationBasic","function":[{"action":"{\"command\":\"DirectionDown\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"DirectionDown","label":"Direction Down"},{"action":"{\"command\":\"DirectionLeft\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"DirectionLeft","label":"Direction Left"},{"action":"{\"command\":\"DirectionRight\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"DirectionRight","label":"Direction Right"},{"action":"{\"command\":\"DirectionUp\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"DirectionUp","label":"Direction Up"},{"action":"{\"command\":\"Enter\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"Select","label":"Select"}]},{"name":"TransportBasic","function":[{"action":"{\"command\":\"UsbStop\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"Stop","label":"Stop"},{"action":"{\"command\":\"UsbPlay\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"Play","label":"Play"},{"action":"{\"command\":\"UsbRewind\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"Rewind","label":"Rewind"},{"action":"{\"command\":\"UsbPause\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"Pause","label":"Pause"},{"action":"{\"command\":\"UsbFastForward\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"FastForward","label":"Fast Forward"}]},{"name":"TransportExtended","function":[{"action":"{\"command\":\"UsbSkipBack\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"SkipBackward","label":"Skip Backward"},{"action":"{\"command\":\"UsbSkipForward\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"SkipForward","label":"Skip Forward"}]},{"name":"NavigationDVD","function":[{"action":"{\"command\":\"Return\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"Return","label":"Return"}]},{"name":"GameType3","function":[{"action":"{\"command\":\"Home\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"Home","label":"Home"}]},{"name":"RadioTuner","function":[{"action":"{\"command\":\"PresetPrev\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"PrevPreset","label":"Prev Preset"},{"action":"{\"command\":\"TuneDown\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"ScanDown","label":"Scan Down"},{"action":"{\"command\":\"TuneUp\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"ScanUp","label":"Scan Up"},{"action":"{\"command\":\"PresetNext\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"NextPreset","label":"Next Preset"},{"action":"{\"command\":\"DirectTune\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"DirectTune","label":"Direct Tune"}]},{"name":"Setup","function":[{"action":"{\"command\":\"Sleep\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"Sleep","label":"Sleep"}]}],"VolumeActivityRole":"47917687","imageKey":"Activity\/3AEBFA09-DE9A-46F5-86FB-AC5E585295FB.png","roles":{"PlayMediaActivityRole":"47917687","VolumeActivityRole":"47917687"},"activityOrder":0,"isMultiZone":false,"baseImageUri":"https:\/\/rcbu-test-ssl-amr.s3.amazonaws.com\/"},{"type":"PowerOff","isAVActivity":false,"label":"PowerOff","id":"-1","activityTypeDisplayName":"Default","controlGroup":[],"sequences":[],"fixit":{"47917687":{"id":"47917687","Power":"Off"},"47918441":{"id":"47918441","Power":"Off"}},"enterActions":[],"roles":[],"rules":[],"icon":"Default","suggestedDisplay":"Default"}],"sequence":[],"device":[{"label":"Onkyo AV Receiver","deviceAddedDate":"\/Date(1507143635950+0000)\/","ControlPort":7,"deviceProfileUri":"svcs.myharmony.com\/res\/device\/47917687-3k\/ZqyVVQlWYjk0gZU7w8yckKgFOigZqVE\/GpixiKPg=","manufacturer":"Onkyo","icon":"5","suggestedDisplay":"DEFAULT","deviceTypeDisplayName":"StereoReceiver","powerFeatures":{"PowerOnActions":[{"__type":"IRPressAction","IRCommandName":"PowerOn","Order":1,"Duration":null,"ActionId":0}],"PowerOffActions":[{"__type":"IRPressAction","IRCommandName":"PowerOff","Order":1,"Duration":null,"ActionId":0}]},"uuid":"aab01371-bd13-71b9-d61a-7109bdb9b40d","Capabilities":[1,5,8,62],"controlGroup":[{"name":"Power","function":[{"action":"{\"command\":\"PowerOff\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"PowerOff","label":"Power Off"},{"action":"{\"command\":\"PowerOn\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"PowerOn","label":"Power On"},{"action":"{\"command\":\"PowerToggle\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"PowerToggle","label":"Power Toggle"}]},{"name":"NumericBasic","function":[{"action":"{\"command\":\"0\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"Number0","label":"0"},{"action":"{\"command\":\"1\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"Number1","label":"1"},{"action":"{\"command\":\"2\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"Number2","label":"2"},{"action":"{\"command\":\"3\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"Number3","label":"3"},{"action":"{\"command\":\"4\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"Number4","label":"4"},{"action":"{\"command\":\"5\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"Number5","label":"5"},{"action":"{\"command\":\"6\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"Number6","label":"6"},{"action":"{\"command\":\"7\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"Number7","label":"7"},{"action":"{\"command\":\"8\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"Number8","label":"8"},{"action":"{\"command\":\"9\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"Number9","label":"9"}]},{"name":"Volume","function":[{"action":"{\"command\":\"Mute\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"Mute","label":"Mute"},{"action":"{\"command\":\"VolumeDown\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"VolumeDown","label":"Volume Down"},{"action":"{\"command\":\"VolumeUp\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"VolumeUp","label":"Volume Up"}]},{"name":"NavigationBasic","function":[{"action":"{\"command\":\"DirectionDown\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"DirectionDown","label":"Direction Down"},{"action":"{\"command\":\"DirectionLeft\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"DirectionLeft","label":"Direction Left"},{"action":"{\"command\":\"DirectionRight\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"DirectionRight","label":"Direction Right"},{"action":"{\"command\":\"DirectionUp\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"DirectionUp","label":"Direction Up"},{"action":"{\"command\":\"Enter\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"Select","label":"Select"}]},{"name":"TransportBasic","function":[{"action":"{\"command\":\"UsbStop\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"Stop","label":"Stop"},{"action":"{\"command\":\"UsbPlay\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"Play","label":"Play"},{"action":"{\"command\":\"UsbRewind\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"Rewind","label":"Rewind"},{"action":"{\"command\":\"UsbPause\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"Pause","label":"Pause"},{"action":"{\"command\":\"UsbFastForward\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"FastForward","label":"Fast Forward"}]},{"name":"TransportExtended","function":[{"action":"{\"command\":\"UsbSkipBack\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"SkipBackward","label":"Skip Backward"},{"action":"{\"command\":\"UsbSkipForward\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"SkipForward","label":"Skip Forward"}]},{"name":"NavigationDVD","function":[{"action":"{\"command\":\"Return\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"Return","label":"Return"}]},{"name":"RadioTuner","function":[{"action":"{\"command\":\"PresetPrev\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"PrevPreset","label":"Prev Preset"},{"action":"{\"command\":\"TuneDown\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"ScanDown","label":"Scan Down"},{"action":"{\"command\":\"TuneUp\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"ScanUp","label":"Scan Up"},{"action":"{\"command\":\"PresetNext\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"NextPreset","label":"Next Preset"},{"action":"{\"command\":\"DirectTune\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"DirectTune","label":"Direct Tune"}]},{"name":"Setup","function":[{"action":"{\"command\":\"Q Setup\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"Setup","label":"Setup"},{"action":"{\"command\":\"Sleep\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"Sleep","label":"Sleep"}]},{"name":"DisplayMode","function":[{"action":"{\"command\":\"Display\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"Display","label":"Display"}]},{"name":"Miscellaneous","function":[{"action":"{\"command\":\"ChannelSelect\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"ChannelSelect","label":"ChannelSelect"},{"action":"{\"command\":\"Dimmer\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"Dimmer","label":"Dimmer"},{"action":"{\"command\":\"HdmiOut1\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"HdmiOut1","label":"HdmiOut1"},{"action":"{\"command\":\"HdmiOut2\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"HdmiOut2","label":"HdmiOut2"},{"action":"{\"command\":\"HdmiOutBoth\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"HdmiOutBoth","label":"HdmiOutBoth"},{"action":"{\"command\":\"HdmiOutWrap\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"HdmiOutWrap","label":"HdmiOutWrap"},{"action":"{\"command\":\"Home\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"Home","label":"Home"},{"action":"{\"command\":\"InputAm\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"InputAm","label":"InputAm"},{"action":"{\"command\":\"InputAux\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"InputAux","label":"InputAux"},{"action":"{\"command\":\"InputBd\\\/Dvd\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"InputBd\/Dvd","label":"InputBd\/Dvd"},{"action":"{\"command\":\"InputCbl\\\/Sat\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"InputCbl\/Sat","label":"InputCbl\/Sat"},{"action":"{\"command\":\"InputExtra1\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"InputExtra1","label":"InputExtra1"},{"action":"{\"command\":\"InputExtra2\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"InputExtra2","label":"InputExtra2"},{"action":"{\"command\":\"InputFm\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"InputFm","label":"InputFm"},{"action":"{\"command\":\"InputGame\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"InputGame","label":"InputGame"},{"action":"{\"command\":\"InputInternetRadio\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"InputInternetRadio","label":"InputInternetRadio"},{"action":"{\"command\":\"InputMusicServer\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"InputMusicServer","label":"InputMusicServer"},{"action":"{\"command\":\"InputNet\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"InputNet","label":"InputNet"},{"action":"{\"command\":\"InputPc\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"InputPc","label":"InputPc"},{"action":"{\"command\":\"InputTv\\\/Cd\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"InputTv\/Cd","label":"InputTv\/Cd"},{"action":"{\"command\":\"InputUsb\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"InputUsb","label":"InputUsb"},{"action":"{\"command\":\"ModeGame\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"ModeGame","label":"ModeGame"},{"action":"{\"command\":\"ModeMovie\\\/Tv\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"ModeMovie\/Tv","label":"ModeMovie\/Tv"},{"action":"{\"command\":\"ModeMusic\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"ModeMusic","label":"ModeMusic"},{"action":"{\"command\":\"ModeStereo\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"ModeStereo","label":"ModeStereo"},{"action":"{\"command\":\"UsbDisplay\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"UsbDisplay","label":"UsbDisplay"},{"action":"{\"command\":\"UsbDown\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"UsbDown","label":"UsbDown"},{"action":"{\"command\":\"UsbLeft\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"UsbLeft","label":"UsbLeft"},{"action":"{\"command\":\"UsbMenu\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"UsbMenu","label":"UsbMenu"},{"action":"{\"command\":\"UsbMode\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"UsbMode","label":"UsbMode"},{"action":"{\"command\":\"UsbRandom\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"UsbRandom","label":"UsbRandom"},{"action":"{\"command\":\"UsbRepeat\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"UsbRepeat","label":"UsbRepeat"},{"action":"{\"command\":\"UsbReturn\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"UsbReturn","label":"UsbReturn"},{"action":"{\"command\":\"UsbRight\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"UsbRight","label":"UsbRight"},{"action":"{\"command\":\"UsbSearch\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"UsbSearch","label":"UsbSearch"},{"action":"{\"command\":\"UsbSelect\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"UsbSelect","label":"UsbSelect"},{"action":"{\"command\":\"UsbSetup\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"UsbSetup","label":"UsbSetup"},{"action":"{\"command\":\"UsbTopMenu\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"UsbTopMenu","label":"UsbTopMenu"},{"action":"{\"command\":\"UsbUp\",\"type\":\"IRCommand\",\"deviceId\":\"47917687\"}","name":"UsbUp","label":"UsbUp"}]}],"DongleRFID":0,"IsKeyboardAssociated":false,"model":"TX-NR515","type":"StereoReceiver","id":"47917687","Transport":1,"isManualPower":"false"},{"label":"Philips TV","deviceAddedDate":"\/Date(1507145506360+0000)\/","ControlPort":7,"deviceProfileUri":"svcs.myharmony.com\/res\/device\/47918441-lfIh\/z53o4veNcTfl0fn2UdX+U5kb2\/\/tOAEcJ+MDm8=","manufacturer":"Philips","icon":"1","suggestedDisplay":"DEFAULT","deviceTypeDisplayName":"Television","powerFeatures":{"PowerOnActions":[{"__type":"IRPressAction","IRCommandName":"PowerOn","Order":0,"Duration":null,"ActionId":0}],"PowerOffActions":[{"__type":"IRPressAction","IRCommandName":"PowerOff","Order":0,"Duration":null,"ActionId":0}]},"Capabilities":[1,2,3,5,8,24],"controlGroup":[{"name":"Power","function":[{"action":"{\"command\":\"PowerOff\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"PowerOff","label":"Power Off"},{"action":"{\"command\":\"PowerOn\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"PowerOn","label":"Power On"},{"action":"{\"command\":\"PowerToggle\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"PowerToggle","label":"Power Toggle"}]},{"name":"NumericBasic","function":[{"action":"{\"command\":\"0\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"Number0","label":"0"},{"action":"{\"command\":\"1\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"Number1","label":"1"},{"action":"{\"command\":\"2\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"Number2","label":"2"},{"action":"{\"command\":\"3\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"Number3","label":"3"},{"action":"{\"command\":\"4\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"Number4","label":"4"},{"action":"{\"command\":\"5\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"Number5","label":"5"},{"action":"{\"command\":\"6\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"Number6","label":"6"},{"action":"{\"command\":\"7\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"Number7","label":"7"},{"action":"{\"command\":\"8\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"Number8","label":"8"},{"action":"{\"command\":\"9\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"Number9","label":"9"}]},{"name":"Volume","function":[{"action":"{\"command\":\"Mute\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"Mute","label":"Mute"},{"action":"{\"command\":\"VolumeDown\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"VolumeDown","label":"Volume Down"},{"action":"{\"command\":\"VolumeUp\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"VolumeUp","label":"Volume Up"}]},{"name":"Channel","function":[{"action":"{\"command\":\"ChannelPrev\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"PrevChannel","label":"Prev Channel"},{"action":"{\"command\":\"ChannelDown\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"ChannelDown","label":"Channel Down"},{"action":"{\"command\":\"ChannelUp\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"ChannelUp","label":"Channel Up"}]},{"name":"NavigationBasic","function":[{"action":"{\"command\":\"DirectionDown\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"DirectionDown","label":"Direction Down"},{"action":"{\"command\":\"DirectionLeft\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"DirectionLeft","label":"Direction Left"},{"action":"{\"command\":\"DirectionRight\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"DirectionRight","label":"Direction Right"},{"action":"{\"command\":\"DirectionUp\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"DirectionUp","label":"Direction Up"},{"action":"{\"command\":\"Ok\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"Select","label":"Select"}]},{"name":"TransportBasic","function":[{"action":"{\"command\":\"Stop\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"Stop","label":"Stop"},{"action":"{\"command\":\"Play\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"Play","label":"Play"},{"action":"{\"command\":\"Rewind\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"Rewind","label":"Rewind"},{"action":"{\"command\":\"Pause\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"Pause","label":"Pause"},{"action":"{\"command\":\"FastForward\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"FastForward","label":"Fast Forward"}]},{"name":"TransportRecording","function":[{"action":"{\"command\":\"Record\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"Record","label":"Record"}]},{"name":"NavigationDVD","function":[{"action":"{\"command\":\"Subtitle\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"Subtitle","label":"Subtitle"},{"action":"{\"command\":\"Back\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"Back","label":"Back"}]},{"name":"Teletext","function":[{"action":"{\"command\":\"Teletext\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"Teletext","label":"Teletext"}]},{"name":"NavigationDSTB","function":[{"action":"{\"command\":\"List\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"List","label":"List"},{"action":"{\"command\":\"Search\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"Search","label":"Search"}]},{"name":"ColoredButtons","function":[{"action":"{\"command\":\"Green\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"Green","label":"Green"},{"action":"{\"command\":\"Red\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"Red","label":"Red"},{"action":"{\"command\":\"Blue\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"Blue","label":"Blue"},{"action":"{\"command\":\"Yellow\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"Yellow","label":"Yellow"}]},{"name":"NavigationExtended","function":[{"action":"{\"command\":\"Guide\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"Guide","label":"Guide"},{"action":"{\"command\":\"Info\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"Info","label":"Info"},{"action":"{\"command\":\"PageDown\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"PageDown","label":"Page Down"},{"action":"{\"command\":\"PageUp\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"PageUp","label":"Page Up"},{"action":"{\"command\":\"Exit\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"Exit","label":"Exit"}]},{"name":"DisplayMode","function":[{"action":"{\"command\":\"Aspect\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"Aspect","label":"Aspect"}]},{"name":"GoogleTVNavigation","function":[{"action":"{\"command\":\"Settings\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"Settings","label":"Settings"}]},{"name":"Miscellaneous","function":[{"action":"{\"command\":\"2D\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"2D","label":"2D"},{"action":"{\"command\":\"3D\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"3D","label":"3D"},{"action":"{\"command\":\"Ambilight\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"Ambilight","label":"Ambilight"},{"action":"{\"command\":\"Help\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"Help","label":"Help"},{"action":"{\"command\":\"Home\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"Home","label":"Home"},{"action":"{\"command\":\"InputHdmi1\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"InputHdmi1","label":"InputHdmi1"},{"action":"{\"command\":\"InputHdmi2\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"InputHdmi2","label":"InputHdmi2"},{"action":"{\"command\":\"InputHdmi3\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"InputHdmi3","label":"InputHdmi3"},{"action":"{\"command\":\"InputHdmi4\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"InputHdmi4","label":"InputHdmi4"},{"action":"{\"command\":\"InputNetwork\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"InputNetwork","label":"InputNetwork"},{"action":"{\"command\":\"InputScart\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"InputScart","label":"InputScart"},{"action":"{\"command\":\"InputUsb\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"InputUsb","label":"InputUsb"},{"action":"{\"command\":\"Multiview\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"Multiview","label":"Multiview"},{"action":"{\"command\":\"Options\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"Options","label":"Options"},{"action":"{\"command\":\"SmartTv\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"SmartTv","label":"SmartTv"},{"action":"{\"command\":\"Source\",\"type\":\"IRCommand\",\"deviceId\":\"47918441\"}","name":"Source","label":"Source"}]}],"DongleRFID":0,"IsKeyboardAssociated":false,"model":"55PUS7909","type":"Television","id":"47918441","Transport":1,"isManualPower":"false"}],"sla":{"latestSLAAcceptedDate":"\/Date(1507143310180+0000)\/","latestSLAAccepted":true},"content":{"contentUserHost":"https:\/\/content.dhg.myharmony.com\/1.0\/User;{userProfileUri}","contentDeviceHost":"https:\/\/content.dhg.myharmony.com\/1.0\/Device;{deviceProfileUri}","contentServiceHost":"https:\/\/content.dhg.myharmony.com\/1.0\/Service\/{providerId}","contentImageHost":"https:\/\/d1tk8oqnnsddt5.cloudfront.net\/1.0\/station\/{stationId}\/image;maxX=40;maxY=40","householdUserProfileUri":"svcs.myharmony.com\/res\/\/household\/8219819-kW+5RwnBrXPTDavh1G1+3G0iI9xvlfCTCSg6XO0thdY=\/user\/default"},"global":{"timeStampHash":"7f4aaf9d-7d1e-4aed-8f36-641da1b123ed9299c3cd-8d2b-4ec7-8ae0-116ae854e9ab\/51447553-5935-4dba-b33d-54d9db33fa5be7d082aa-46de-49e1-99ec-c6364a7151f15331a1bd-56c2-4c90-ad1e-b3dfeb2d41c311254819Harmony+Huben-USFIMS-e2bee8d2-a4c6-484c-b629-6ddab8bf656b0137866762254445643europe%2fhelsinkiTrue17308367611;4aa53e6141ece12b039f90109bbbf539","locale":"en-US"}}]]></oa></iq>
*/

// Scan XML token stream to find next StartElement.
func nextStart(p *xml.Decoder) (xml.StartElement, error) {
	for {
		t, err := p.Token()
		if err != nil || t == nil {
			return xml.StartElement{}, err
		}
		switch t := t.(type) {
		case xml.StartElement:
			return t, nil
		}
	}
}

// Scan XML token stream for next element and save into val.
// If val == nil, allocate new element based on proto map.
// Either way, return val.
func nextFullElement(p *xml.Decoder) (xml.Name, interface{}, error) {
	// Read start element to find out what type we want.
	se, err := nextStart(p)
	if err != nil {
		return xml.Name{}, nil, err
	}

	// Put it in an interface and allocate one.
	var nv interface{}
	switch se.Name.Space + " " + se.Name.Local { // FIXME: use prettyXmlName
	case "http://etherx.jabber.org/streams features":
		nv = &streamFeatures{}
	case "urn:ietf:params:xml:ns:xmpp-sasl success":
		nv = &saslSuccess{}
	case "urn:ietf:params:xml:ns:xmpp-sasl failure":
		nv = &saslFailure{}
	case "jabber:client iq":
		nv = &clientIQ{}
	/*
		case nsStream + " error":
			nv = &streamError{}
		case nsTLS + " starttls":
			nv = &tlsStartTLS{}
		case nsTLS + " proceed":
			nv = &tlsProceed{}
		case nsTLS + " failure":
			nv = &tlsFailure{}
		case "urn:ietf:params:xml:ns:xmpp-sasl mechanisms":
			nv = &saslMechanisms{}
		case "urn:ietf:params:xml:ns:xmpp-sasl challenge":
			nv = ""
		case "urn:ietf:params:xml:ns:xmpp-sasl response":
			nv = ""
		case "urn:ietf:params:xml:ns:xmpp-sasl abort":
			nv = &saslAbort{}
		case nsBind + " bind":
			nv = &bindBind{}
		case "jabber:client message":
			nv = &clientMessage{}
		case "jabber:client presence":
			nv = &clientPresence{}
		case "jabber:client error":
			nv = &clientError{}
	*/
	default:
		return xml.Name{}, nil, errors.New("unexpected XMPP message " +
			se.Name.Space + " <" + se.Name.Local + "/>")
	}

	// Unmarshal into that storage.
	if err = p.DecodeElement(nv, &se); err != nil {
		return xml.Name{}, nil, err
	}

	return se.Name, nv, err
}

type HarmonyHubConnection struct {
	addr       string
	conn       net.Conn
	connected  bool
	xmlDecoder *xml.Decoder
}

func (h *HarmonyHubConnection) connectAndDoTheDance() error {
	conn, err := net.Dial("tcp", h.addr)
	if err != nil {
		return err
	}

	xmlDecoder := xml.NewDecoder(conn)
	xmlDecoder.CharsetReader = charset.NewReaderLabel

	h.conn = conn
	h.xmlDecoder = xmlDecoder

	if err := h.InitAndAuthenticate(); err != nil {
		return err
	}

	// does not actually go to that hostname/central service, but instead just the end device..
	// (bad name for stream recipient)
	if err := h.StartStreamTo("connect.logitech.com"); err != nil {
		return err
	}

	if err := h.Bind(); err != nil {
		return err
	}

	log.Printf("HarmonyHubConnection: connected")

	return nil
}

func NewHarmonyHubConnection(addr string, stopper *stopper.Stopper) *HarmonyHubConnection {
	defer stopper.Done()

	harmonyHubConnection := &HarmonyHubConnection{
		addr:      addr,
		connected: false,
	}

	keepaliveTicker := time.NewTicker(20 * time.Second)

	// hub disconnects us in 60s unless we send application level
	// keepalives (line breaks). TCP level keepalives didn't seem to work.
	go func() {
		for {
			select {
			case <-stopper.ShouldStop:
				log.Println("harmony: stopping")
				keepaliveTicker.Stop()

				if harmonyHubConnection.connected {
					harmonyHubConnection.EndStream()
				}
				return
			case <-time.After(1 * time.Second):
				if !harmonyHubConnection.connected {
					log.Printf("harmony: connecting ...")

					if err := harmonyHubConnection.connectAndDoTheDance(); err != nil {
						log.Printf("harmony: failed to connect: %s", err.Error())
					} else {
						harmonyHubConnection.connected = true
					}
				}
			case <-keepaliveTicker.C:
				// not using Send() to suppress debug logging,
				// and this is not application level stuff anyway
				if _, err := harmonyHubConnection.conn.Write([]byte("\n")); err != nil {
					log.Printf("harmony: failed to send keepalive newline: %s", err.Error())
					break
				}
			}
		}
	}()

	return harmonyHubConnection
}

func prettyXmlName(name xml.Name) string {
	return name.Space + " " + name.Local
}

func (x *HarmonyHubConnection) InitAndAuthenticate() error {
	openingMsg := `<stream:stream xmlns="jabber:client" xmlns:stream="http://etherx.jabber.org/streams" version="1.0" to="x.com">`

	if err := x.Send(openingMsg); err != nil {
		return err
	}

	// We expect the server to start a <stream>.
	streamStartEl, err := nextStart(x.xmlDecoder)
	if err != nil {
		return err
	}
	if prettyXmlName(streamStartEl.Name) != "http://etherx.jabber.org/streams stream" {
		return fmt.Errorf("expected <stream> but got %s", prettyXmlName(streamStartEl.Name))
	}

	// Now we're in the stream and can use Unmarshal.
	// Next message should be <features> to tell us authentication options.
	// See section 4.6 in RFC 3920.
	// f := new(streamFeatures)
	elName, el, err := nextFullElement(x.xmlDecoder)
	if err != nil {
		return err
	}

	if prettyXmlName(elName) != "http://etherx.jabber.org/streams features" {
		return errors.New("expecting <features>")
	}

	featuresEl := el.(*streamFeatures)

	if len(featuresEl.Mechanisms.Mechanism) != 1 {
		return errors.New("expecting one mechanism")
	}

	firstMechanism := featuresEl.Mechanisms.Mechanism[0]

	if firstMechanism != "PLAIN" {
		return errors.New("expecting PLAIN mechanism")
	}

	// don't know why it's OK to log in as guest to the hub...
	plainAuthContent := saslAuthString("guest@x.com", "guest", "guest")

	authElXml, _ := xml.Marshal(saslAuth{
		Mechanism: "PLAIN",
		Content:   plainAuthContent,
	})

	if err := x.Send(string(authElXml) + "\n"); err != nil {
		return err
	}

	// Next message should be either success or failure.
	authRespElName, authResp, err := nextFullElement(x.xmlDecoder)
	if err != nil {
		return err
	}
	switch v := authResp.(type) {
	case *saslSuccess:
	case *saslFailure:
		errorMessage := v.Text
		if errorMessage == "" {
			// v.Any is type of sub-element in failure,
			// which gives a description of what failed if there was no text element
			errorMessage = v.Any.Local
		}
		return errors.New("auth failure: " + errorMessage)
	default:
		return errors.New(fmt.Sprintf("expected <success> or <failure>, got %s", prettyXmlName(authRespElName)))
	}

	return nil
}

func (x *HarmonyHubConnection) Send(msg string) error {
	log.Printf("harmony: > %s", msg)

	_, err := x.conn.Write([]byte(msg))

	if err != nil {
		x.connected = false
	}

	return err
}

func (x *HarmonyHubConnection) StartStreamTo(to string) error {
	openStreamMsg := `<stream:stream xmlns="jabber:client" xmlns:stream="http://etherx.jabber.org/streams" version="1.0" to="` + to + `">`
	if err := x.Send(openStreamMsg); err != nil {
		return err
	}

	// We expect the server to start a <stream>.
	streamStartEl, err := nextStart(x.xmlDecoder)
	if err != nil {
		return err
	}
	if prettyXmlName(streamStartEl.Name) != "http://etherx.jabber.org/streams stream" {
		return fmt.Errorf("expected <stream> but got %s", prettyXmlName(streamStartEl.Name))
	}

	// not interested in contents of the features element
	featuresElName, _, err := nextFullElement(x.xmlDecoder)
	if err != nil {
		return err
	}

	if prettyXmlName(featuresElName) != "http://etherx.jabber.org/streams features" {
		return errors.New("expecting <features>")
	}

	return nil
}

func (x *HarmonyHubConnection) EndStream() error {
	endStreamMsg := `</stream:stream>`
	if err := x.Send(endStreamMsg); err != nil {
		return err
	}

	// TODO: expect back stream end

	return nil
}

func (x *HarmonyHubConnection) Bind() error {
	// https://xmpp.org/rfcs/rfc3920.html#bind

	// remember, the id attribute is our own generated identifier
	bindEl := `<iq type="set" id="bind123"><bind xmlns="urn:ietf:params:xml:ns:xmpp-bind"><resource>gatorade</resource></bind></iq>`

	if err := x.Send(bindEl); err != nil {
		return err
	}

	// expecting this back:
	// <iq id='bind123' type='result'><bind xmlns='urn:ietf:params:xml:ns:xmpp-bind'><jid>1111/gatorade</jid></bind></iq>

	iqName1, _, err := nextFullElement(x.xmlDecoder)
	if err != nil {
		return err
	}

	if prettyXmlName(iqName1) != "jabber:client iq" {
		return errors.New("expecting <iq/>")
	}

	return nil
}

func (x *HarmonyHubConnection) Recv() {
	xmlName, _, err := nextFullElement(x.xmlDecoder)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("< %s", prettyXmlName(xmlName))
}

func (x *HarmonyHubConnection) HoldAndRelease(deviceId string, commandName string) error {
	if !x.connected {
		return errNotConnected
	}

	// <iq> = infoquery
	// attribute "id" is randomly generated, but not required

	pressPayload := fmt.Sprintf(`action={"command"::"%s","type"::"IRCommand","deviceId"::"%s"}:status=press`, commandName, deviceId)
	releasePayload := fmt.Sprintf(`action={"command"::"%s","type"::"IRCommand","deviceId"::"%s"}:status=release`, commandName, deviceId)

	pressEl := fmt.Sprintf(`<iq type="get"><oa xmlns="connect.logitech.com" mime="vnd.logitech.harmony/vnd.logitech.harmony.engine?holdAction">%s</oa></iq>`, pressPayload)
	releaseEl := fmt.Sprintf(`<iq type="get"><oa xmlns="connect.logitech.com" mime="vnd.logitech.harmony/vnd.logitech.harmony.engine?holdAction">%s</oa></iq>`, releasePayload)

	if err := x.Send(pressEl + releaseEl); err != nil {
		return err
	}

	// expecting two emmpty <iq> elements back

	iqName1, _, err := nextFullElement(x.xmlDecoder)
	if err != nil {
		return err
	}

	if prettyXmlName(iqName1) != "jabber:client iq" {
		return errors.New("expecting first <iq/>")
	}

	iqName2, _, err2 := nextFullElement(x.xmlDecoder)
	if err2 != nil {
		return err2
	}

	if prettyXmlName(iqName2) != "jabber:client iq" {
		return errors.New("expecting second <iq/>")
	}

	return nil
}

func saslAuthString(email string, login string, pwd string) string {
	return base64.StdEncoding.EncodeToString(
		[]byte(fmt.Sprintf("%s\x00%s\x00%s", email, login, pwd)))
}

// types

type saslAuth struct {
	XMLName   xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-sasl auth"`
	Mechanism string   `xml:"mechanism,attr"`
	Content   string   `xml:",chardata"`
}

// RFC 3920  C.1  Streams name space
type streamFeatures struct {
	XMLName    xml.Name `xml:"http://etherx.jabber.org/streams features"`
	StartTLS   *tlsStartTLS
	Mechanisms saslMechanisms
	Bind       bindBind
	Session    bool
}

// RFC 3920  C.3  TLS name space
type tlsStartTLS struct {
	XMLName  xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-tls starttls"`
	Required *string  `xml:"required"`
}

// RFC 3920  C.4  SASL name space
type saslMechanisms struct {
	XMLName   xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-sasl mechanisms"`
	Mechanism []string `xml:"mechanism"`
}

// RFC 3920  C.5  Resource binding name space
type bindBind struct {
	XMLName  xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-bind bind"`
	Resource string
	Jid      string `xml:"jid"`
}

type saslSuccess struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-sasl success"`
}

type saslFailure struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-sasl failure"`
	Any     xml.Name `xml:",any"`
	Text    string   `xml:"text"`
}

type clientIQ struct { // info/query
	XMLName xml.Name `xml:"jabber:client iq"`
	From    string   `xml:"from,attr"`
	ID      string   `xml:"id,attr"`
	To      string   `xml:"to,attr"`
	Type    string   `xml:"type,attr"` // error, get, result, set
	Query   []byte   `xml:",innerxml"`
	Error   clientError
	Bind    bindBind
}

type clientError struct {
	XMLName xml.Name `xml:"jabber:client error"`
	Code    string   `xml:",attr"`
	Type    string   `xml:",attr"`
	Any     xml.Name
	Text    string
}
