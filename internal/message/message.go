package message

import "time"

// Entity represents a text formatting entity (bold, italic, link, etc).
type Entity struct {
	Type   string `yaml:"type" json:"type"`
	Offset int    `yaml:"offset" json:"offset"`
	Length int    `yaml:"length" json:"length"`
	URL    string `yaml:"url,omitempty" json:"url,omitempty"`
}

// Message represents a single message from any provider.
type Message struct {
	From         string     `yaml:"from"`
	Provider     string     `yaml:"provider"`
	Channel      string     `yaml:"channel"`
	Date         time.Time  `yaml:"date"`
	ID           string     `yaml:"id"`
	ReplyTo      string     `yaml:"reply_to,omitempty"`
	ReplyToBody  string     `yaml:"reply_to_body,omitempty"`
	Quote        string     `yaml:"quote,omitempty"`
	MediaType    string     `yaml:"media_type,omitempty"`
	MediaFileID  string     `yaml:"media_file_id,omitempty"`
	MediaURL     string     `yaml:"media_url,omitempty"`
	Caption      string     `yaml:"caption,omitempty"`
	ForwardFrom  string     `yaml:"forward_from,omitempty"`
	ForwardDate  *time.Time `yaml:"forward_date,omitempty"`
	EditDate     *time.Time `yaml:"edit_date,omitempty"`
	ThreadID     string     `yaml:"thread_id,omitempty"`
	MediaGroupID string     `yaml:"media_group_id,omitempty"`
	Entities     []Entity   `yaml:"entities,omitempty"`
	Body         string     `yaml:"-"`

	// Transient fields for media download (not serialized).
	DownloadURL string `yaml:"-" json:"-"`
	MediaExt    string `yaml:"-" json:"-"`
}
