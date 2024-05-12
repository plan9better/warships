package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func CheckValidCoords(coords string) bool {
	// Length (3 and 4 because of trailing '\n' )
	if len(coords) < 3 || len(coords) > 4 {
		return false
	}

	// Check if between A-J
	if int(coords[0]) < 65 || int(coords[0]) > 74 {
		return false
	}

	// Check if between 0-9
	if int(coords[1]) < 48 || int(coords[1]) > 57 {
		return false
	}

	// if 2 digit number check if second digit == 0
	if len(coords) == 4 && int(coords[2]) != 48 {
		return false
	}

	return true

}

func PromptString(info string, def string) string {

	str := def
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Enter %s (default is '%s'): ", info, def)
	text, _ := reader.ReadBytes('\n')
	if len(text) == 1 {
		fmt.Printf("No %s provided, going with the default\n", info)
		return def
	}
	str = string(text)
	str = strings.TrimSuffix(str, "\n")

	// log.Println("promtstring returning: ", str)
	return str
}
