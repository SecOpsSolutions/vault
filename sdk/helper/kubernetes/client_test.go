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
	e.TestGetPod(t)
	e.TestGetPodNotFound(t)
	e.TestUpdatePodTags(t)
	e.TestUpdatePodTagsNotFound(t)
}

type env struct {
	client        LightWeightClient
	currentLabels map[string]string
}

func (e *env) TestGetPod(t *testing.T) {
	if err := e.client.GetPod(TestNamespace, TestPodname); err != nil {
		t.Fatal(err)
	}
}

func (e *env) TestGetPodNotFound(t *testing.T) {
	err := e.client.GetPod(TestNamespace, "no-exist")
	if err == nil {
		t.Fatal("expected error because pod is unfound")
	}
	if err != ErrNotFound {
		t.Fatalf("expected %q but received %q", ErrNotFound, err)
	}
}

func (e *env) TestUpdatePodTags(t *testing.T) {
	if err := e.client.UpdatePodTags(TestNamespace, TestPodname, &Tag{
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

func (e *env) TestUpdatePodTagsNotFound(t *testing.T) {
	err := e.client.UpdatePodTags(TestNamespace, "no-exist", &Tag{
		Key:   "fizz",
		Value: "buzz",
	})
	if err == nil {
		t.Fatal("expected error because pod is unfound")
	}
	if err != ErrNotFound {
		t.Fatalf("expected %q but received %q", ErrNotFound, err)
	}
}
