// internal/app/objects.go
package app

import (
	"encoding/json"
	"os"
)

func WriteJSONFile(path string, data interface{}) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func GetJSONFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}
