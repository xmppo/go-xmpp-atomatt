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

		break
	}

	return newXMPP(jid, stream), nil
}

func startClient(stream *Stream, jid JID) error {

	s := fmt.Sprintf(
		"<stream:stream from='%s' to='%s' version='1.0' xml:lang='en' xmlns='jabber:client' xmlns:stream='http://etherx.jabber.org/streams'>",
		jid,
		jid.Domain)
	if err := stream.Send(s); err != nil {
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
	
	x := fmt.Sprintf(
		"<auth xmlns='urn:ietf:params:xml:ns:xmpp-sasl' mechanism='PLAIN'>%s</auth>",
		saslEncodePlain(user, password))
	if err := stream.Send(x); err != nil {
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
}

type mechanisms struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-sasl mechanisms"`
	Mechanisms []string `xml:"mechanism"`
}

type tlsStartTLS struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-tls starttls"`
	Required *tlsStartTLSRequired `xml:"required"`
}

type tlsStartTLSRequired struct {
}

type saslSuccess struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-sasl success"`
}

type saslFailure struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-sasl failure"`
	Reason xml.Name `xml:",any"`
}
