package systemd

import "testing"

func TestUnitNameAndUser(t *testing.T) {
	if UnitName("api") != "velo-api.service" {
		t.Fatalf("unit: %s", UnitName("api"))
	}
	if LegacyUnitName("api") != "deploy-api.service" {
		t.Fatalf("legacy: %s", LegacyUnitName("api"))
	}
	if AppUser("api") != "velo-api" {
		t.Fatalf("user: %s", AppUser("api"))
	}
}
