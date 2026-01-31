package ui

import (
	"fmt"
	"os"
)

func ExitWithError(msg string, err error) {
	if err != nil {
		fmt.Println(StatusError(fmt.Sprintf("%s: %v", msg, err)))
	} else {
		fmt.Println(StatusError(msg))
	}
	os.Exit(1)
}
