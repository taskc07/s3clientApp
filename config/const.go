package config

import _ "embed"

const ServerPort = "4040"
const TRUE = "true"
const FALSE = "false"

//go:embed file.pdf
var PdfBytes []byte
