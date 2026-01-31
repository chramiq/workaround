package config

import (
	_ "embed"
)

//go:embed worker.js
var DefaultWorkerScript []byte

//go:embed useragents.txt
var DefaultUserAgents []byte