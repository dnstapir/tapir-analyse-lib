package common

const NATSHEADER_KEY_IDENTIFIER = "DNSTAPIR-Key-Identifier"
const NATSHEADER_KEY_THUMBPRINT = "DNSTAPIR-Key-Thumbprint"
const NATSHEADER_MSG_SCHEMA = "DNSTAPIR-Message-Schema"

const NATS_WILDCARD = "*"
const NATS_GLOB = ">"
const NATS_DELIM = "."

var NATSHEADERS_DNSTAPIR_ALL = []string{
	NATSHEADER_KEY_IDENTIFIER,
	NATSHEADER_KEY_THUMBPRINT,
	NATSHEADER_MSG_SCHEMA,
}

type NatsMsg struct {
	Headers map[string]string
	Subject string
	Data    []byte
}
