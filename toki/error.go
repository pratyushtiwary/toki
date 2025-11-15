package toki

import (
	"os"

	"github.com/pratyushtiwary/toki/log"
)

func Check(e error, prefix string) {
	if e != nil {
		log.Error(prefix+": %s\n", e.Error())
		os.Exit(1)
	}
}
