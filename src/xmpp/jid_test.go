package xmpp

import "testing"

func TestBare(t *testing.T) {
	if (JID{"node", "domain", "resource"}).Bare() != "node@domain" {
		t.FailNow()
	}
}

func TestFull(t *testing.T) {
	if (JID{"node", "domain", "resource"}).Full() != "node@domain/resource" {
		t.FailNow()
	}
	if (JID{"node", "domain", ""}).Full() != "node@domain" {
		t.FailNow()
	}
	if (JID{"", "domain", ""}).Full() != "domain" {
		t.FailNow()
	}
}

func TestParseJID(t *testing.T) {
	jid, _ := ParseJID("node@domain/resource")
	if jid != (JID{"node", "domain", "resource"}) {
		t.FailNow()
	}
	jid, _ = ParseJID("node@domain")
	if jid != (JID{"node", "domain", ""}) {
		t.FailNow()
	}
	jid, _ = ParseJID("domain")
	if jid != (JID{"", "domain", ""}) {
		t.FailNow()
	}
}
