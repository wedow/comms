package embeddocs

import _ "embed"

//go:embed telegram-setup.md
var TelegramSetupDoc []byte
