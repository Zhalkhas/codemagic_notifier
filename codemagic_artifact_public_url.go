package codemagic_notifier

import "time"

type CodeMagicArtifactPublicUrl struct {
	Url       string    `json:"url"`
	ExpiresAt time.Time `json:"expiresAt"`
}
