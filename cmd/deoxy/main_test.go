package main

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func TestVersionOutput(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	main()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if output != "deoxy v0.1.0\n" {
		t.Errorf("expected 'deoxy v0.1.0\\n', got %q", output)
	}
}

func TestVersionVariable(t *testing.T) {
	if version != "0.1.0" {
		t.Errorf("expected version '0.1.0', got %q", version)
	}
}
