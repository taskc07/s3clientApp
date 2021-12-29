package config

import (
	_ "embed"
)

const ServerPort = "4040"
const MaxGoRoutinesExecution = 300

//go:embed file.pdf
var PdfBytes []byte

var NullBytes = make([]byte, 0)
