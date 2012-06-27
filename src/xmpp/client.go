package xmpp

import (
	"crypto/tls"
	"encoding/xml"
	"errors"
	"fmt"
	"log"
)

type Client struct {
	JID JID
	stream *Stream
}

type ClientConfig struct {
	NoTLS bool
	InsecureSkipVerify bool
}

func NewClient(jid JID, password string, config *ClientConfig) (*Client, error) {

	stream, err := NewStream(jid.Domain + ":5222")
	if err != nil {
		return nil, err
	}

	if err := stream.Send("<?xml version='1.0'?>\n"); err != nil {
		return nil, err
	}

	for {
		// Send stream start.
		s := fmt.Sprintf(
			"<stream:stream from='%s' to='%s' version='1.0' xml:lang='en' xmlns='jabber:client' xmlns:stream='http://etherx.jabber.org/streams'>",
			jid,
			jid.Domain)
		if err := stream.Send(fmt.Sprintf(s)); err != nil {
			return nil, err
		}

		// Receive stream start.
		if _, err := stream.Next(&xml.Name{nsStream, "stream"}); err != nil {
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
			if err := authenticate(stream, f.Mechanisms.Mechanisms, jid.Local, password); err != nil {
				return nil, err
			}
			continue // Restart
		}

		break
	}

	return &Client{jid, stream}, nil
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
