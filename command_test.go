package main

import (
	"fmt"
	"testing"
)

func TestParseCommands(t *testing.T) {
	cmd, err := parseCommands([]string{"gator", "login", "jan"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmd.Name != "login" {
		t.Errorf("Name = %q, want %q", cmd.Name, "login")
	}
	if len(cmd.Args) != 1 || cmd.Args[0] != "jan" {
		t.Errorf("Args = %#v, want [jan]", cmd.Args)
	}

	fmt.Println(cmd)
}
