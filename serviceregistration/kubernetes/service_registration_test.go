package kubernetes

import (
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/vault/sdk/helper/kubernetes"
	sr "github.com/hashicorp/vault/serviceregistration"
)

var testVersion = "version 1"

func TestServiceRegistration(t *testing.T) {
	currentLabels, closeFunc := kubernetes.TestServer(t)
	defer closeFunc()

	if len(currentLabels) != 0 {
		t.Fatalf("expected 0 current labels but have %d: %s", len(currentLabels), currentLabels)
	}
	shutdownCh := make(chan struct{})
	config := map[string]string{
		"namespace":    kubernetes.TestNamespace,
		"service_name": kubernetes.TestServiceName,
	}
	logger := hclog.NewNullLogger()
	state := &sr.State{
		VaultVersion:         testVersion,
		IsInitialized:        true,
		IsSealed:             true,
		IsActive:             true,
		IsPerformanceStandby: true,
	}
	reg, err := NewServiceRegistration(shutdownCh, config, logger, state, "")
	if err != nil {
		t.Fatal(err)
	}

	// Test initial state.
	if len(currentLabels) != 5 {
		t.Fatalf("expected 5 current labels but have %d: %s", len(currentLabels), currentLabels)
	}
	if currentLabels[labelVaultVersion] != testVersion {
		t.Fatalf("expected %q but received %q", testVersion, currentLabels[labelVaultVersion])
	}
	if currentLabels[labelActive] != toString(true) {
		t.Fatalf("expected %q but received %q", toString(true), currentLabels[labelVaultVersion])
	}
	if currentLabels[labelSealed] != toString(true) {
		t.Fatalf("expected %q but received %q", toString(true), currentLabels[labelVaultVersion])
	}
	if currentLabels[labelPerfStandby] != toString(true) {
		t.Fatalf("expected %q but received %q", toString(true), currentLabels[labelVaultVersion])
	}
	if currentLabels[labelInitialized] != toString(true) {
		t.Fatalf("expected %q but received %q", toString(true), currentLabels[labelVaultVersion])
	}

	// Test NotifyActiveStateChange.
	if err := reg.NotifyActiveStateChange(false); err != nil {
		t.Fatal(err)
	}
	if currentLabels[labelActive] != toString(false) {
		t.Fatalf("expected %q but received %q", toString(false), currentLabels[labelActive])
	}
	if err := reg.NotifyActiveStateChange(true); err != nil {
		t.Fatal(err)
	}
	if currentLabels[labelActive] != toString(true) {
		t.Fatalf("expected %q but received %q", toString(true), currentLabels[labelActive])
	}

	// Test NotifySealedStateChange.
	if err := reg.NotifySealedStateChange(false); err != nil {
		t.Fatal(err)
	}
	if currentLabels[labelSealed] != toString(false) {
		t.Fatalf("expected %q but received %q", toString(false), currentLabels[labelSealed])
	}
	if err := reg.NotifySealedStateChange(true); err != nil {
		t.Fatal(err)
	}
	if currentLabels[labelSealed] != toString(true) {
		t.Fatalf("expected %q but received %q", toString(true), currentLabels[labelSealed])
	}

	// Test NotifyPerformanceStandbyStateChange.
	if err := reg.NotifyPerformanceStandbyStateChange(false); err != nil {
		t.Fatal(err)
	}
	if currentLabels[labelPerfStandby] != toString(false) {
		t.Fatalf("expected %q but received %q", toString(false), currentLabels[labelPerfStandby])
	}
	if err := reg.NotifyPerformanceStandbyStateChange(true); err != nil {
		t.Fatal(err)
	}
	if currentLabels[labelPerfStandby] != toString(true) {
		t.Fatalf("expected %q but received %q", toString(true), currentLabels[labelPerfStandby])
	}

	// Test NotifyInitializedStateChange.
	if err := reg.NotifyInitializedStateChange(false); err != nil {
		t.Fatal(err)
	}
	if currentLabels[labelInitialized] != toString(false) {
		t.Fatalf("expected %q but received %q", toString(false), currentLabels[labelInitialized])
	}
	if err := reg.NotifyInitializedStateChange(true); err != nil {
		t.Fatal(err)
	}
	if currentLabels[labelInitialized] != toString(true) {
		t.Fatalf("expected %q but received %q", toString(true), currentLabels[labelInitialized])
	}
}
