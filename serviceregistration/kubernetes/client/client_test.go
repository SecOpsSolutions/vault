package client

import (
	"os"
	"testing"

	"github.com/hashicorp/go-hclog"
	kubetest "github.com/hashicorp/vault/serviceregistration/kubernetes/testing"
)

func TestClient(t *testing.T) {
	testState, testConf, closeFunc := kubetest.Server(t)
	defer closeFunc()

	Scheme = testConf.ClientScheme
	TokenFile = testConf.PathToTokenFile
	RootCAFile = testConf.PathToRootCAFile
	if err := os.Setenv(EnvVarKubernetesServiceHost, testConf.ServiceHost); err != nil {
		t.Fatal(err)
	}
	if err := os.Setenv(EnvVarKubernetesServicePort, testConf.ServicePort); err != nil {
		t.Fatal(err)
	}

	client, err := New(hclog.Default(), make(chan struct{}))
	if err != nil {
		t.Fatal(err)
	}
	e := &env{
		client:    client,
		testState: testState,
	}
	e.TestGetPod(t)
	e.TestGetPodNotFound(t)
	e.TestUpdatePodTags(t)
	e.TestUpdatePodTagsNotFound(t)
}

type env struct {
	client    *Client
	testState *kubetest.State
}

func (e *env) TestGetPod(t *testing.T) {
	pod, err := e.client.GetPod(kubetest.ExpectedNamespace, kubetest.ExpectedPodName)
	if err != nil {
		t.Fatal(err)
	}
	if pod.Metadata.Name != "shell-demo" {
		t.Fatalf("expected %q but received %q", "shell-demo", pod.Metadata.Name)
	}
}

func (e *env) TestGetPodNotFound(t *testing.T) {
	_, err := e.client.GetPod(kubetest.ExpectedNamespace, "no-exist")
	if err == nil {
		t.Fatal("expected error because pod is unfound")
	}
	if _, ok := err.(*ErrNotFound); !ok {
		t.Fatalf("expected *ErrNotFound but received %T", err)
	}
}

func (e *env) TestUpdatePodTags(t *testing.T) {
	if err := e.client.PatchPod(kubetest.ExpectedNamespace, kubetest.ExpectedPodName, &Patch{
		Operation: Add,
		Path:      "/metadata/labels/fizz",
		Value:     "buzz",
	}); err != nil {
		t.Fatal(err)
	}
	if e.testState.NumPatches() != 1 {
		t.Fatalf("expected 1 label but received %+v", e.testState)
	}
	if e.testState.Get("/metadata/labels/fizz")["value"] != "buzz" {
		t.Fatalf("expected buzz but received %q", e.testState.Get("fizz")["value"])
	}
	if e.testState.Get("/metadata/labels/fizz")["op"] != "add" {
		t.Fatalf("expected add but received %q", e.testState.Get("fizz")["op"])
	}
}

func (e *env) TestUpdatePodTagsNotFound(t *testing.T) {
	err := e.client.PatchPod(kubetest.ExpectedNamespace, "no-exist", &Patch{
		Operation: Add,
		Path:      "/metadata/labels/fizz",
		Value:     "buzz",
	})
	if err == nil {
		t.Fatal("expected error because pod is unfound")
	}
	if _, ok := err.(*ErrNotFound); !ok {
		t.Fatalf("expected *ErrNotFound but received %T", err)
	}
}
