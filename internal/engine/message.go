package engine

import (
	"encoding/json"
	"time"
)

// Message represents a message passed between nodes
type Message struct {
	Payload  interface{}            `json:"payload"`
	Topic    string                 `json:"topic"`
	Headers  map[string]string      `json:"headers"`
	Metadata map[string]interface{} `json:"metadata"`
	SourceID string                 `json:"sourceId"`
	MsgID    string                 `json:"msgId"`
	Timestamp time.Time             `json:"timestamp"`
}

// NewMessage creates a new message with the given payload
func NewMessage(payload interface{}, topic string) *Message {
	return &Message{
		Payload:   payload,
		Topic:     topic,
		Headers:   make(map[string]string),
		Metadata:  make(map[string]interface{}),
		MsgID:     generateUUID(),
		Timestamp: time.Now(),
	}
}

// Clone creates a deep copy of the message
func (m *Message) Clone() *Message {
	clone := &Message{
		Topic:     m.Topic,
		SourceID:  m.SourceID,
		MsgID:     m.MsgID,
		Timestamp: m.Timestamp,
		Headers:   make(map[string]string),
		Metadata:  make(map[string]interface{}),
	}
	
	// Copy headers
	for k, v := range m.Headers {
		clone.Headers[k] = v
	}
	
	// Copy metadata
	for k, v := range m.Metadata {
		clone.Metadata[k] = v
	}
	
	// Deep copy payload
	if m.Payload != nil {
		payloadBytes, err := json.Marshal(m.Payload)
		if err == nil {
			var clonedPayload interface{}
			if err = json.Unmarshal(payloadBytes, &clonedPayload); err == nil {
				clone.Payload = clonedPayload
			} else {
				// Fallback to shallow copy if deep copy fails
				clone.Payload = m.Payload
			}
		} else {
			// Fallback to shallow copy if deep copy fails
			clone.Payload = m.Payload
		}
	}
	
	return clone
}

// SetPayload sets the message payload
func (m *Message) SetPayload(payload interface{}) {
	m.Payload = payload
}

// SetTopic sets the message topic
func (m *Message) SetTopic(topic string) {
	m.Topic = topic
}

// SetHeader sets a header value
func (m *Message) SetHeader(key, value string) {
	m.Headers[key] = value
}

// GetHeader gets a header value
func (m *Message) GetHeader(key string) (string, bool) {
	value, exists := m.Headers[key]
	return value, exists
}

// SetMetadata sets a metadata value
func (m *Message) SetMetadata(key string, value interface{}) {
	m.Metadata[key] = value
}

// GetMetadata gets a metadata value
func (m *Message) GetMetadata(key string) (interface{}, bool) {
	value, exists := m.Metadata[key]
	return value, exists
}

// ToJSON converts the message to JSON
func (m *Message) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}

// FromJSON populates the message from JSON
func (m *Message) FromJSON(data []byte) error {
	return json.Unmarshal(data, m)
}

// generateUUID generates a simple UUID
// In a production environment, use a proper UUID library
func generateUUID() string {
	return time.Now().Format("20060102150405.000000000") + "-" + 
		   randomHex(4) + "-" + 
		   randomHex(4) + "-" + 
		   randomHex(4) + "-" + 
		   randomHex(12)
}

// randomHex generates a random hex string of the given length
func randomHex(length int) string {
	const hexChars = "0123456789abcdef"
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = hexChars[time.Now().UnixNano()%16]
		time.Sleep(1 * time.Nanosecond) // Ensure different values
	}
	return string(result)
}
