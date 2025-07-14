// Package util provides utility functions for the gh-models extension.
package util

import (
	"fmt"
	"io"
	"os"
)

// WriteToOut writes a message to the given io.Writer.
func WriteToOut(out io.Writer, message string) {
	_, err := io.WriteString(out, message)
	if err != nil {
		fmt.Println("Error writing message:", err)
	}
}

// Ptr returns a pointer to the given value.
func Ptr[T any](value T) *T {
	return &value
}

// ReadFile reads the contents of a file and returns it as a byte slice.
func ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

// WriteFile writes data to a file.
func WriteFile(filename string, data []byte) error {
	return os.WriteFile(filename, data, 0644)
}
