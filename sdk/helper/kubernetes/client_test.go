package kubernetes

import (
	"testing"
)

func TestClient(t *testing.T) {
	currentLabels, closeFunc := TestServer(t)
	defer closeFunc()

	client, err := NewLightWeightClient()
	if err != nil {
		t.Fatal(err)
	}
	e := &env{
		client:        client,
		currentLabels: currentLabels,
	}
	e.TestGetService(t)
	e.TestGetServiceNotFound(t)
	e.TestUpdateServiceSelectors(t)
	e.TestUpdateServiceSelectorsNotFound(t)
}

type env struct {
	client        LightWeightClient
	currentLabels map[string]string
}

func (e *env) TestGetService(t *testing.T) {
	if err := e.client.GetService(TestNamespace, TestServiceName); err != nil {
		t.Fatal(err)
	}
}

func (e *env) TestGetServiceNotFound(t *testing.T) {
	err := e.client.GetService(TestNamespace, "no-exist")
	if err == nil {
		t.Fatal("expected error because service is unfound")
	}
	if err != ErrNotFound {
		t.Fatalf("expected %q but received %q", ErrNotFound, err)
	}
}

func (e *env) TestUpdateServiceSelectors(t *testing.T) {
	if err := e.client.UpdateServiceSelectors(TestNamespace, TestServiceName, &Selector{
		Key:   "fizz",
		Value: "buzz",
	}); err != nil {
		t.Fatal(err)
	}
	if len(e.currentLabels) != 1 {
		t.Fatalf("expected 1 label but received %q", e.currentLabels)
	}
	if e.currentLabels["fizz"] != "buzz" {
		t.Fatalf("expected buzz but received %q", e.currentLabels["fizz"])
	}
}

func (e *env) TestUpdateServiceSelectorsNotFound(t *testing.T) {
	err := e.client.UpdateServiceSelectors(TestNamespace, "no-exist", &Selector{
		Key:   "fizz",
		Value: "buzz",
	})
	if err == nil {
		t.Fatal("expected error because service is unfound")
	}
	if err != ErrNotFound {
		t.Fatalf("expected %q but received %q", ErrNotFound, err)
	}
}
