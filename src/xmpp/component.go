package xmpp

import (
	"crypto/sha1"
	"encoding/xml"
	"errors"
	"fmt"
	"log"
)

// Create a component XMPP stream.
func NewComponentStream(addr string, jid JID, secret string) (*Stream, error) {

	stream, err := NewStream(addr)
	if err != nil {
		return nil, err
	}

	streamId, err := startComponent(stream, jid)
	if err != nil {
		return nil, err
	}

	if err := handshake(stream, streamId, secret); err != nil {
		return nil, err
	}

	return stream, nil
}

func startComponent(stream *Stream, jid JID) (string, error) {

	s := fmt.Sprintf(
		"<stream:stream xmlns='jabber:component:accept' xmlns:stream='http://etherx.jabber.org/streams' to='%s'>",
		jid)
	if err := stream.Send(s); err != nil {
		return "", err
	}

	streamId := ""
	if e, err := stream.Next(&xml.Name{nsStream, "stream"}); err != nil {
		return "", err
	} else {
		// Find the stream id.
		for _, attr := range e.Attr {
			if attr.Name.Local == "id" {
				streamId = attr.Value
				break
			}
		}
		if streamId == "" {
			return "", errors.New("Missing stream id")
		}
	}

	return streamId, nil
}

func handshake(stream *Stream, streamId, secret string) error {

	hash := sha1.New()
	hash.Write([]byte(streamId))
	hash.Write([]byte(secret))

	// Send handshake.
	s := fmt.Sprintf("<handshake>%x</handshake>", hash.Sum(nil))
	log.Println(s)
	if err := stream.Send(s); err != nil {
		return err
	}

	// Get handshake response.
	if _, err := stream.Next(&xml.Name{"jabber:component:accept", "handshake"}); err != nil {
		return err
	}

	return nil
}
