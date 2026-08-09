package main

import (
	"testing"
)

func TestVersionDefaults(t *testing.T) {
	if version == "" {
		t.Errorf("version should not be empty")
	}
}
