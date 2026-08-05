package data

import "embed"

//go:embed languages/*.json
var LanguagesFS embed.FS

//go:embed quotes/*.json
var QuotesFS embed.FS
