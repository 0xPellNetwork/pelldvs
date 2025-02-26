package tools

import (
	"encoding/json"
	"fmt"
)

func PrettyPrint(msg string, data any) {
	b, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		fmt.Println("Error marshalling data:", err)
	}
	fmt.Println()
	fmt.Println(msg)
	fmt.Println(string(b))
	fmt.Println()
}
