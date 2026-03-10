package embeddocs

import _ "embed"

//go:embed telegram-setup.md
var TelegramSetupDoc []byte

//go:embed prime.md
var PrimeDoc []byte
