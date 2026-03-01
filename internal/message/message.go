package message

import "time"

// Message represents a single message from any provider.
type Message struct {
	From     string    `yaml:"from"`
	Provider string    `yaml:"provider"`
	Channel  string    `yaml:"channel"`
	Date     time.Time `yaml:"date"`
	ID       string    `yaml:"id"`
	Body     string    `yaml:"-"`
}
