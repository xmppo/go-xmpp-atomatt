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

// Create a client XMPP over the stream.
func NewClientXMPP(stream *Stream, jid JID, password string, config *ClientConfig) (*XMPP, error) {

	if config == nil {
		config = &ClientConfig{}
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
			log.Println("Start TLS")
			if err := startTLS(stream, config); err != nil {
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
			xml.Attr{xml.Name{"", "version"}, "1.0"},
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

func startTLS(stream *Stream, config *ClientConfig) error {

	if err := stream.Send(&tlsStart{}); err != nil {
		return err
	}

	p := tlsProceed{}
	if err := stream.Decode(&p); err != nil {
		return err
	}

	tlsConfig := tls.Config{InsecureSkipVerify: config.InsecureSkipVerify}
	return stream.UpgradeTLS(&tlsConfig)
}

type tlsStart struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-tls starttls"`
}

type tlsProceed struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-tls proceed"`
}

func authenticate(stream *Stream, mechanisms []string, user, password string) error {
	for _, handler := range authHandlers {
		if !stringSliceContains(mechanisms, handler.Mechanism) {
			continue
		}
		if err := handler.Fn(stream, user, password); err == nil {
			log.Printf("Authentication (%s) successful", handler.Mechanism)
			return nil
		}
	}
	return errors.New("No supported SASL mechanism found.")
}

type authHandler struct {
	Mechanism string
	Fn func(*Stream, string, string) error
}

var authHandlers = []authHandler{
		authHandler{"PLAIN", authenticatePlain},
	}

func authenticatePlain(stream *Stream, user, password string) error {
	auth := saslAuth{Mechanism: "PLAIN", Text: saslEncodePlain(user, password)}
	if err := stream.Send(&auth); err != nil {
		return err
	}

	se, err := stream.Next(nil)
	if err != nil {
		return err
	}
	switch se.Name.Local {
	case "success":
		if err := stream.Skip(); err != nil {
			return err
		}
	case "failure":
		f := new(saslFailure)
		stream.DecodeElement(f, se)
		return errors.New(fmt.Sprintf("Authentication failed: %s", f.Reason.Local))
	}

	return nil
}

type saslAuth struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-sasl auth"`
	Mechanism string `xml:"mechanism,attr"`
	Text string `xml:",chardata"`
}

func bindResource(stream *Stream, jid JID) (JID, error) {

	req := Iq{Id: "foo", Type: "set"}
	if jid.Resource == "" {
		req.PayloadEncode(bindIq{})
	} else {
		req.PayloadEncode(bindIq{Resource: jid.Resource})
	}
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

type saslFailure struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-sasl failure"`
	Reason xml.Name `xml:",any"`
}

// BUG(matt): Don't use "foo" as the <iq/> id during resource binding.
