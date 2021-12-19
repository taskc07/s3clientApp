package config

import _ "embed"

const ServerPort = "4040"

//go:embed file.pdf
var PdfBytes []byte
