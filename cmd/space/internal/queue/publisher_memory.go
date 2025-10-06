package queue

import "log"

type InMemoryPublisher struct{}

func NewInMemoryPublisher() *InMemoryPublisher { return &InMemoryPublisher{} }

func (p *InMemoryPublisher) Publish(topic, key string, payload []byte) error {
	log.Printf("event topic=%s key=%s payload=%dB", topic, key, len(payload))
	return nil
}
