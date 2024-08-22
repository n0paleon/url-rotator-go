package pkg

import (
	"github.com/teris-io/shortid"
	"regexp"
)

func GenerateShortID() string {
	id := shortid.MustGenerate()
	id = regexp.MustCompile(`[^a-zA-Z0-9]`).ReplaceAllString(id, "")

	return id
}
