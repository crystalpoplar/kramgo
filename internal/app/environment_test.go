package app

import (
	"fmt"
	"io"
)

func testCreateRequiredDirectories(w io.Writer) error {
	_, err := fmt.Fprintln(w, "kramgo is ready")
	for _, dir := range requiredDirectories {
		err := createDirectoryIfNotExists(fmt.Sprintf("%s/%s", homeDirectory, dir))
		if err != nil {
			return err
		}
	}
	return err
}
