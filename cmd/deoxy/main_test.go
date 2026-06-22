package main

import "testing"

func TestVersionVariable(t *testing.T) {
	if version != "0.1.0" {
		t.Errorf("expected version '0.1.0', got %q", version)
	}
}

func TestVersionNotEmpty(t *testing.T) {
	if version == "" {
		t.Error("version variable should not be empty")
	}
}
