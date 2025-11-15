package toki

import (
	"fmt"
	"os"
)

func Check(e error) {
	if e != nil {
		fmt.Printf("ERROR: %s\n", e.Error())
		os.Exit(1)
	}
}
