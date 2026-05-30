// Package jobs defines the payloads exchanged over Redis Streams between the
// webhook ingestion service and the orchestration workers.
package jobs

// InboundJob is enqueued per inbound contact message and consumed by the
// orchestration worker. All values are strings (Redis Stream fields).
type InboundJob struct {
	CompanyID      string `redis:"company_id"`
	ChannelID      string `redis:"channel_id"`
	ConversationID string `redis:"conversation_id"`
	MessageID      string `redis:"message_id"`
	ExternalID     string `redis:"external_id"`
	Instance       string `redis:"instance"`
	RemoteJID      string `redis:"remote_jid"`
	MessageType    string `redis:"message_type"` // conversation | audioMessage | ...
	Content        string `redis:"content"`
}

// ToMap renders the job as a Redis Stream field map.
func (j InboundJob) ToMap() map[string]any {
	return map[string]any{
		"company_id":      j.CompanyID,
		"channel_id":      j.ChannelID,
		"conversation_id": j.ConversationID,
		"message_id":      j.MessageID,
		"external_id":     j.ExternalID,
		"instance":        j.Instance,
		"remote_jid":      j.RemoteJID,
		"message_type":    j.MessageType,
		"content":         j.Content,
	}
}

// FromMap parses a Redis Stream field map into an InboundJob.
func FromMap(m map[string]any) InboundJob {
	get := func(k string) string {
		if v, ok := m[k]; ok {
			if s, ok := v.(string); ok {
				return s
			}
		}
		return ""
	}
	return InboundJob{
		CompanyID:      get("company_id"),
		ChannelID:      get("channel_id"),
		ConversationID: get("conversation_id"),
		MessageID:      get("message_id"),
		ExternalID:     get("external_id"),
		Instance:       get("instance"),
		RemoteJID:      get("remote_jid"),
		MessageType:    get("message_type"),
		Content:        get("content"),
	}
}
