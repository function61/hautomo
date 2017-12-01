package main

import (
	"encoding/xml"
	"errors"
	"fmt"
	"golang.org/x/net/html/charset"
	"log"
	"net"
	// "github.com/mattn/go-xmpp"
)

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
	conn       net.Conn
	xmlDecoder *xml.Decoder
}

func NewHarmonyHubConnection(addr string) *HarmonyHubConnection {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	xmlDecoder := xml.NewDecoder(conn)
	xmlDecoder.CharsetReader = charset.NewReaderLabel

	return &HarmonyHubConnection{
		conn:       conn,
		xmlDecoder: xmlDecoder,
	}
}

func prettyXmlName(name xml.Name) string {
	return name.Space + " " + name.Local
}

func (x *HarmonyHubConnection) InitAndAuthenticate() error {
	openingMsg := `<stream:stream xmlns="jabber:client" xmlns:stream="http://etherx.jabber.org/streams" version="1.0" to="x.com">`

	x.Send(openingMsg)

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

	// guest
	authCreds := "Z3Vlc3RAeC5jb20AZ3Vlc3QAZ3Vlc3Q="
	authMsg := fmt.Sprintf("<auth xmlns='urn:ietf:params:xml:ns:xmpp-sasl' mechanism='PLAIN'>%s</auth>\n", authCreds)

	x.Send(authMsg)

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
	log.Printf("> %s", msg)

	// TODO: error
	fmt.Fprintf(x.conn, msg)

	return nil
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

// types

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
