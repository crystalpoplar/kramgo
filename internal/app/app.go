package app

import (
	"fmt"
	"io"
)

func Run(w io.Writer) error {
	_, err := fmt.Fprintln(w, "kramgo is ready")
	return err
}
