package gt

import (
	"errors"
)

const (
	JSON DecodeFormat = "json"
	YAML DecodeFormat = "yaml"
	// BODY 可以解析 string，[]byte
	BODY DecodeFormat = "body"
)

var (
	DecoderTypeNotSupport = errors.New("decoder type not support")
)

type DecodeFormat string
