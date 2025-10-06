package queue

import "log"

type MemoryPublisher struct{}

func NewMemoryPublisher() *MemoryPublisher { return &MemoryPublisher{} }

func (p *MemoryPublisher) Publish(topic string, key string, payload []byte) error {
	log.Printf("event topic=%s key=%s payload=%d bytes", topic, key, len(payload))
	return nil
}
