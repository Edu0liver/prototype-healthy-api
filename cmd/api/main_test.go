package main

import (
	"testing"

	"go.uber.org/fx"
)

// TestAppWiring validates the full fx dependency graph (every provider/invoke is
// resolvable, no missing dependencies or cycles) without opening DB/Redis
// connections or running OnStart hooks. This is the offline equivalent of
// booting the app: it fails fast on misconfigured wiring (e.g. a module that
// needs the billing service before it is provided).
func TestAppWiring(t *testing.T) {
	if err := fx.ValidateApp(options()); err != nil {
		t.Fatalf("fx app wiring invalid: %v", err)
	}
}
