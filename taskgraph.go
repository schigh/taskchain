package taskgraph

import (
	"io"
	"io/ioutil"
	"log"
	"os"
)

var (
	lg *log.Logger
)

func init() {
	os.Setenv("DEBUG", "TRUE")
	var writer io.Writer
	if os.Getenv("DEBUG") != "" {
		writer = os.Stderr
	} else {
		writer = ioutil.Discard
	}
	lg = log.New(writer, "📌 ", log.LUTC)
}
