package xmpp

import (
	"bytes"
	"encoding/xml"
	"fmt"
)

// XMPP <iq/> stanza.
type Iq struct {
	XMLName xml.Name `xml:"iq"`
	Id      string   `xml:"id,attr"`
	Type    string   `xml:"type,attr"`
	To      string   `xml:"to,attr,omitempty"`
	From    string   `xml:"from,attr,omitempty"`
	Payload string   `xml:",innerxml"`
	Error   *Error   `xml:"error"`
}

// Encode the value to an XML string and set as the payload. See xml.Marshal
// for how the value is encoded.
func (iq *Iq) PayloadEncode(v interface{}) error {
	bytes, err := xml.Marshal(v)
	if err != nil {
		return err
	}
	iq.Payload = string(bytes)
	return nil
}

// Decode the payload (an XML string) into the given value. See xml.Unmarshal
// for how the value is decoded.
func (iq *Iq) PayloadDecode(v interface{}) error {
	return xml.Unmarshal([]byte(iq.Payload), v)
}

// Return the name of the payload element.
func (iq *Iq) PayloadName() (name xml.Name) {
	dec := xml.NewDecoder(bytes.NewBufferString(iq.Payload))
	tok, err := dec.Token()
	if err != nil {
		return
	}
	start, ok := tok.(xml.StartElement)
	if !ok {
		return
	}
	return start.Name
}

// Create a response Iq. The Id is kept, To and From are reversed, Type is set
// to the given value.
func (iq *Iq) Response(type_ string) *Iq {
	return &Iq{Id: iq.Id, Type: type_, From: iq.To, To: iq.From}
}

// XMPP <message/> stanza.
type Message struct {
	XMLName xml.Name `xml:"message"`
	Id      string   `xml:"id,attr,omitempty"`
	Type    string   `xml:"type,attr,omitempty"`
	To      string   `xml:"to,attr,omitempty"`
	From    string   `xml:"from,attr,omitempty"`
	Subject string   `xml:"subject,omitempty"`
	Body    string   `xml:"body,omitempty"`
}

// XMPP <presence/> stanza.
type Presence struct {
	XMLName xml.Name `xml:"presence"`
	Id      string   `xml:"id,attr,omitempty"`
	Type    string   `xml:"type,attr,omitempty"`
	To      string   `xml:"to,attr,omitempty"`
	From    string   `xml:"from,attr,omitempty"`
}

// XMPP <error/>. May occur as a top-level stanza or embedded in another
// stanza, e.g. an <iq type="error"/>.
type Error struct {
	XMLName xml.Name `xml:"error"`
	Type    string   `xml:"type,attr"`
	Payload string   `xml:",innerxml"`
}

func (e Error) Error() string {
	if text := e.Text(); text == "" {
		return fmt.Sprintf("[%s] %s", e.Type, e.Condition().Local)
	} else {
		return fmt.Sprintf("[%s] %s, %s", e.Type, e.Condition().Local, text)
	}
	panic("unreachable")
}

type errorText struct {
	XMLName xml.Name
	Text    string `xml:",chardata"`
}

// Create a new Error instance using the args as the payload.
func NewError(errorType string, condition ErrorCondition, text string) *Error {

	// Build payload.
	buf := new(bytes.Buffer)
	writeXMLStartElement(buf, &xml.StartElement{
		Name: xml.Name{"", condition.Local},
		Attr: []xml.Attr{
			{xml.Name{"", "xmlns"}, condition.Space},
		},
	})
	writeXMLEndElement(buf, &xml.EndElement{Name: xml.Name{"", condition.Local}})
	enc := xml.NewEncoder(buf)
	if text != "" {
		enc.Encode(errorText{xml.Name{condition.Space, "text"}, text})
	}

	return &Error{Type: errorType, Payload: string(buf.Bytes())}
}

// Return the error text from the payload, or "" if not present.
func (e Error) Text() string {
	dec := xml.NewDecoder(bytes.NewBufferString(e.Payload))
	next := startElementIter(dec)
	for start := next(); start != nil; {
		if start.Name.Local == "text" {
			text := errorText{}
			dec.DecodeElement(&text, start)
			return text.Text
		}
		dec.Skip()
		start = next()
	}
	return ""
}

// Return the error condition from the payload.
func (e Error) Condition() ErrorCondition {
	dec := xml.NewDecoder(bytes.NewBufferString(e.Payload))
	next := startElementIter(dec)
	for start := next(); start != nil; {
		if start.Name.Local != "text" && (start.Name.Space == nsErrorStanzas || start.Name.Space == nsErrorStreams) {
			return ErrorCondition(start.Name)
		}
		dec.Skip()
		start = next()
	}
	return ErrorCondition{}
}

// Error condition.
type ErrorCondition xml.Name

// Stanza errors.
var (
	FeatureNotImplemented = ErrorCondition{nsErrorStanzas, "feature-not-implemented"}
)
