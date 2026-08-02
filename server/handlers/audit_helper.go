package handlers

import "ababilx-sso/services/audit"

// identityAuditEvent builds an audit.Event from the nullable
// ip/user-agent pointers clientContext returns, so call sites don't
// each repeat the nil-dereference dance.
func identityAuditEvent(userID, event string, ipHash, userAgent *string, meta map[string]any) audit.Event {
	e := audit.Event{ActorUserID: userID, Event: event, Meta: meta}
	if ipHash != nil {
		e.IPHash = *ipHash
	}
	if userAgent != nil {
		e.UserAgent = *userAgent
	}
	return e
}
