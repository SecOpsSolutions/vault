package main

// This code builds a minimal binary of the lightweight kubernetes
// client and exposes it for manual testing.
// The intention is that the binary can be built and dropped into
// a Kube environment like this:
// https://kubernetes.io/docs/tasks/debug-application-cluster/get-shell-running-container/
// Then, commands can be run to test its API calls.
// The above commands are intended to be run inside an instance of
// minikube that has been started.
// After building this binary, place it in the container like this:
// $ kubectl cp kubeclient /shell-demo:/
// At first you may get 403's, which can be resolved using this:
// https://github.com/fabric8io/fabric8/issues/6840#issuecomment-307560275
//
// Example calls:
// 		./kubeclient -call='get-service' -namespace='default' -service-name='shell-demo'
// 		./kubeclient -call='update-service-tags' -namespace='default' -service-name='shell-demo' -selectors='fizz:buzz,foo:bar'

import (
	"flag"
	"fmt"
	"strings"

	"github.com/hashicorp/vault/sdk/helper/kubernetes"
)

var callToMake string
var selectorsToAdd string
var namespace string
var serviceName string

func init() {
	flag.StringVar(&callToMake, "call", "", `the call to make: 'get-service' or 'update-service-selectors'`)
	flag.StringVar(&selectorsToAdd, "selectors", "", `if call is "update-service-selectors", that selectors to update like so: "fizz:buzz,foo:bar"`)
	flag.StringVar(&namespace, "namespace", "", "the namespace to use")
	flag.StringVar(&serviceName, "service-name", "", "the service name to use")
}

func main() {
	flag.Parse()

	client, err := kubernetes.NewLightWeightClient()
	if err != nil {
		panic(err)
	}

	switch callToMake {
	case "get-service":
		if err := client.GetService(namespace, serviceName); err != nil {
			panic(err)
		}
		return
	case "update-service-selectors":
		tagPairs := strings.Split(selectorsToAdd, ",")
		var selectors []*kubernetes.Selector
		for _, tagPair := range tagPairs {
			fields := strings.Split(tagPair, ":")
			if len(fields) != 2 {
				panic(fmt.Errorf("unable to split %s from selectors provided of %s", fields, selectorsToAdd))
			}
			selectors = append(selectors, &kubernetes.Selector{
				Key:   fields[0],
				Value: fields[1],
			})
		}
		if err := client.UpdateServiceSelectors(namespace, serviceName, selectors...); err != nil {
			panic(err)
		}
		return
	default:
		panic(fmt.Errorf(`unsupported call provided: %q`, callToMake))
	}
}
