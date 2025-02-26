package dvs

import (
	"encoding/json"
	"os"
)

func decodeJSONFromFile(filepath string, data any) error {
	input, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}
	err = json.Unmarshal(input, data)
	return err
}
