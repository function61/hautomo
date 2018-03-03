package harmonyhub

import (
	"encoding/xml"
	"github.com/function61/eventhorizon/util/ass"
	"testing"
)

func TestSaslAuth(t *testing.T) {
	auth := saslAuth{
		Mechanism: "PLAIN",
		Content:   saslAuthString("guest@x.com", "guest", "guest"),
	}

	asXml, _ := xml.Marshal(auth)

	ass.EqualString(t, string(asXml), `<auth xmlns="urn:ietf:params:xml:ns:xmpp-sasl" mechanism="PLAIN">Z3Vlc3RAeC5jb20AZ3Vlc3QAZ3Vlc3Q=</auth>`)
}
