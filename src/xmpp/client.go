package xmpp

import (
	"crypto/tls"
	"encoding/xml"
	"errors"
	"fmt"
	"log"
)

// Config structure used to create a new XMPP client connection.
type ClientConfig struct {
	// Don't upgrade the connection to TLS, even if the server supports it. If
	// the server *requires* TLS then this option is ignored.
	NoTLS bool

	// Skip verification of the server's certificate chain. Probably only
	// useful during development.
	InsecureSkipVerify bool
}

// Create a client XMPP stream.
func NewClientXMPP(jid JID, password string, config *ClientConfig) (*XMPP, error) {

	stream, err := NewStream(jid.Domain + ":5222")
	if err != nil {
		return nil, err
	}

	for {

		if err := startClient(stream, jid); err != nil {
			return nil, err
		}

		// Read features.
		f := new(features)
		if err := stream.Decode(f); err != nil {
			return nil, err
		}

		// TLS?
		if f.StartTLS != nil && (f.StartTLS.Required != nil || !config.NoTLS) {
			tlsConfig := tls.Config{InsecureSkipVerify: config.InsecureSkipVerify}
			if err := stream.UpgradeTLS(&tlsConfig); err != nil {
				return nil, err
			}
			continue // Restart
		}

		// Authentication
		if f.Mechanisms != nil {
			log.Println("Authenticating")
			if err := authenticate(stream, f.Mechanisms.Mechanisms, jid.Node, password); err != nil {
				return nil, err
			}
			continue // Restart
		}

		// Bind resource.
		if f.Bind != nil {
			log.Println("Binding resource.")
			boundJID, err := bindResource(stream, jid)
			if err != nil {
				return nil, err
			}
			jid = boundJID
		}

		break
	}

	return newXMPP(jid, stream), nil
}

func startClient(stream *Stream, jid JID) error {

	start := xml.StartElement{
		xml.Name{"stream", "stream"},
		[]xml.Attr{
			xml.Attr{xml.Name{"", "xmlns"}, "jabber:client"},
			xml.Attr{xml.Name{"xmlns", "stream"}, "http://etherx.jabber.org/streams"},
			xml.Attr{xml.Name{"", "from"}, jid.Full()},
			xml.Attr{xml.Name{"", "to"}, jid.Domain},
		},
	}

	if err := stream.SendStart(&start); err != nil {
		return err
	}

	if _, err := stream.Next(&xml.Name{nsStream, "stream"}); err != nil {
		return err
	}

	return nil
}

func authenticate(stream *Stream, mechanisms []string, user, password string) error {

	log.Println("authenticate, mechanisms=", mechanisms)

	if !stringSliceContains(mechanisms, "PLAIN") {
		return errors.New("Only PLAIN supported for now")
	}

	return authenticatePlain(stream, user, password)
}

func authenticatePlain(stream *Stream, user, password string) error {
	
	auth := saslAuth{Mechanism: "PLAIN", Message: saslEncodePlain(user, password)}
	if err := stream.Send(&auth); err != nil {
		return err
	}

	if se, err := stream.Next(nil); err != nil {
		return err
	} else {
		if se.Name.Local == "failure" {
			f := new(saslFailure)
			stream.DecodeElement(f, se)
			return errors.New(fmt.Sprintf("Authentication failed: %s", f.Reason.Local))
		}
	}

	return nil
}

type saslAuth struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-sasl auth"`
	Mechanism string `xml:"mechanism,attr"`
	Message string `xml:",innerxml"`
}

func bindResource(stream *Stream, jid JID) (JID, error) {
	if jid.Resource == "" {
		return bindResourceServer(stream)
	}
	return bindResourceClient(stream, jid)
}

func bindResourceClient(stream *Stream, jid JID) (JID, error) {

	req := Iq{Id: "foo", Type: "set"}
	req.PayloadEncode(bindIq{Resource: jid.Resource})
	if err := stream.Send(req); err != nil {
		return JID{}, err
	}

	resp := Iq{}
	err := stream.Decode(&resp)
	if err != nil {
		return JID{}, err
	}
	bindResp := bindIq{}
	resp.PayloadDecode(&bindResp)

	boundJID, err := ParseJID(bindResp.JID)
	return boundJID, nil
}

func bindResourceServer(stream *Stream) (JID, error) {
	panic("bindResourceServer not implemented")
}

type bindIq struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-bind bind"`
	Resource string `xml:"resource,omitempty"`
	JID string `xml:"jid,omitempty"`
}

func stringSliceContains(l []string, m string) bool {
	for _, i := range l {
		if i == m {
			return true
		}
	}
	return false
}

type features struct {
	XMLName xml.Name `xml:"http://etherx.jabber.org/streams features"`
	StartTLS *tlsStartTLS `xml:"starttls"`
	Mechanisms *mechanisms `xml:"mechanisms"`
	Bind *bind `xml:"bind"`
}

type bind struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-bind bind"`
	Required *required `xml:"required"`
}

type mechanisms struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-sasl mechanisms"`
	Mechanisms []string `xml:"mechanism"`
}

type tlsStartTLS struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-tls starttls"`
	Required *required `xml:"required"`
}

type required struct {}

type saslSuccess struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-sasl success"`
}

type saslFailure struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-sasl failure"`
	Reason xml.Name `xml:",any"`
}
