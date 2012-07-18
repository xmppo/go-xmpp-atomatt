package xmpp

import (
	"crypto/sha1"
	"encoding/xml"
	"errors"
	"fmt"
)

// Create a component XMPP connection over the stream.
func NewComponentXMPP(stream *Stream, jid JID, secret string) (*XMPP, error) {

	streamId, err := startComponent(stream, jid)
	if err != nil {
		return nil, err
	}

	if err := handshake(stream, streamId, secret); err != nil {
		return nil, err
	}

	return newXMPP(jid, stream), nil
}

func startComponent(stream *Stream, jid JID) (string, error) {

	start := xml.StartElement{
		xml.Name{"stream", "stream"},
		[]xml.Attr{
			xml.Attr{xml.Name{"", "xmlns"}, nsComponentAccept},
			xml.Attr{xml.Name{"xmlns", "stream"}, nsStreams},
			xml.Attr{xml.Name{"", "to"}, jid.Full()},
		},
	}

	var streamId string

	if rstart, err := stream.SendStart(&start); err != nil {
		return "", err
	} else {
		if rstart.Name != (xml.Name{nsStreams, "stream"}) {
			return "", fmt.Errorf("unexpected start element: %s", rstart.Name)
		}
		// Find the stream id.
		for _, attr := range rstart.Attr {
			if attr.Name.Local == "id" {
				streamId = attr.Value
				break
			}
		}
	}

	if streamId == "" {
		return "", errors.New("Missing stream id")
	}

	return streamId, nil
}

func handshake(stream *Stream, streamId, secret string) error {

	hash := sha1.New()
	hash.Write([]byte(streamId))
	hash.Write([]byte(secret))

	// Send handshake.
	handshake := saslHandshake{Value: fmt.Sprintf("%x", hash.Sum(nil))}
	if err := stream.Send(&handshake); err != nil {
		return err
	}

	// Get handshake response.
	if start, err := stream.Next(); err != nil {
		return err
	} else {
		if start.Name != (xml.Name{nsComponentAccept, "handshake"}) {
			return fmt.Errorf("Expected <handshake/>, for %s", start.Name)
		}
	}
	if err := stream.Skip(); err != nil {
		return err
	}

	return nil
}

type saslHandshake struct {
	XMLName xml.Name `xml:"jabber:component:accept handshake"`
	Value   string   `xml:",chardata"`
}
