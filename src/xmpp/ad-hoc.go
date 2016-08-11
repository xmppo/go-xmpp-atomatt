package xmpp

import (
	"encoding/xml"
)

const (
	NodeAdHocCommand = "http://jabber.org/protocol/commands"

	ActionAdHocExecute = "execute"
	ActionAdHocNext    = "next"
	ActionAdHocCancel  = "cancel"

	StatusAdHocExecute   = "executing"
	StatusAdHocCompleted = "completed"
	StatusAdHocCanceled  = "canceled"

	TypeAdHocForm   = "form"
	TypeAdHocResult = "result"
	TypeAdHocSubmit = "submit"

	TypeAdHocListSingle = "list-single"
	TypeAdHocListMulti  = "list-multi"

	TypeAdHocNoteInfo    = "info"
	TypeAdHocNoteWarning = "warn"
	TypeAdHocNoteError   = "error"

	TypeAdHocFieldListMulti  = "list-multi"
	TypeAdHocFieldListSingle = "list-single"
	TypeAdHocFieldTextSingle = "text-single"
	TypeAdHocFieldJidSingle  = "jid-single"
	TypeAdHocFieldTextPrivate = "text-private"
)

type AdHocCommand struct {
	XMLName   xml.Name   `xml:"http://jabber.org/protocol/commands command"`
	Node      string     `xml:"node,attr"`
	Action    string     `xml:"action,attr"`
	SessionID string     `xml:"sessionid,attr"`
	Status    string     `xml:"status,attr"`
	XForm     AdHocXForm `xml:"x"`
	Note      AdHocNote  `xml:"note,omitempty"`
}

type AdHocXForm struct {
	XMLName      xml.Name     `xml:"jabber:x:data x"`
	Type         string       `xml:"type,attr"`
	Title        string       `xml:"title"`
	Instructions string       `xml:"instructions"`
	Fields       []AdHocField `xml:"field"`
}

type AdHocField struct {
	Var     string             `xml:"var,attr"`
	Label   string             `xml:"label,attr"`
	Type    string             `xml:"type,attr"`
	Options []AdHocFieldOption `xml:"option"`
	Value   string             `xml:"value,omitempty"`
}

type AdHocFieldOption struct {
	Value string `xml:"value"`
}

type AdHocNote struct {
	Type  string `xml:"type,attr"`
	Value string `xml:",innerxml"`
}
