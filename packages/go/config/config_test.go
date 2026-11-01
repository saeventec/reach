package config

import "testing"

func TestPortUsesTheDefaultWhenNotSet(t *testing.T) {
	// Setenv sets the variable for this test only, and restores it after.
	t.Setenv("PORT", "")

	got := Port()

	if got != "8080" {
		t.Errorf("got %q, want %q", got, "8080")
	}
}

func TestPortReadsTheEnvironment(t *testing.T) {
	t.Setenv("PORT", "9090")

	got := Port()

	if got != "9090" {
		t.Errorf("got %q, want %q", got, "9090")
	}
}

func TestDatabaseURLMustBeSet(t *testing.T) {
	t.Setenv("DATABASE_URL", "")

	_, err := DatabaseURL()

	if err == nil {
		t.Error("got no error, want one: DATABASE_URL has no default")
	}
}
