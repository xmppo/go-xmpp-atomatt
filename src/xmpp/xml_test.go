package xmpp

import (
	"bytes"
	"encoding/xml"
	"testing"
)

func TestWriteNameLocal(t *testing.T) {
	buf := new(bytes.Buffer)
	writeXMLName(buf, xml.Name{"", "foo"})
	if buf.String() != "foo" {
		t.Fail()
	}
}

func TestWriteName(t *testing.T) {
	buf := new(bytes.Buffer)
	writeXMLName(buf, xml.Name{"space", "foo"})
	if buf.String() != "space:foo" {
		t.Fail()
	}
}

func TestWriteAttrLocal(t *testing.T) {
	buf := new(bytes.Buffer)
	writeXMLAttr(buf, xml.Attr{xml.Name{"", "foo"}, "bar"})
	if buf.String() != "foo='bar'" {
		t.Fail()
	}
}

func TestWriteAttr(t *testing.T) {
	buf := new(bytes.Buffer)
	writeXMLAttr(buf, xml.Attr{xml.Name{"space", "foo"}, "bar"})
	if buf.String() != "space:foo='bar'" {
		t.Fail()
	}
}
