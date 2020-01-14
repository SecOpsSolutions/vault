package kubernetes

import (
	"fmt"

	log "github.com/hashicorp/go-hclog"
	kubeHlpr "github.com/hashicorp/vault/sdk/helper/kubernetes"
	sr "github.com/hashicorp/vault/serviceregistration"
)

const (
	defaultNamespace        = "default" // This is Kubernetes' default namespace.
	defaultVaultServiceName = "vault"

	labelVaultVersion = "vault-version"
	labelActive       = "vault-ha-active"
	labelSealed       = "vault-ha-sealed"
	labelPerfStandby  = "vault-ha-perf-standby"
	labelInitialized  = "vault-ha-initialized"
)

func NewServiceRegistration(shutdownCh <-chan struct{}, config map[string]string, logger log.Logger, state *sr.State, _ string) (sr.ServiceRegistration, error) {
	client, err := kubeHlpr.NewLightWeightClient()
	if err != nil {
		return nil, err
	}

	// Perform an initial labelling of Vault as it starts up.
	namespace := config["namespace"]
	if namespace == "" {
		namespace = defaultNamespace
	}

	serviceName := config["service_name"]
	if serviceName == "" {
		serviceName = defaultVaultServiceName
	}

	// Verify that the service exists and our configuration looks good.
	if err := client.GetService(namespace, serviceName); err != nil {
		return nil, err
	}

	// Label the service with our initial values.
	selectors := []*kubeHlpr.Selector{
		{Key: labelVaultVersion, Value: state.VaultVersion},
		{Key: labelActive, Value: toString(state.IsActive)},
		{Key: labelSealed, Value: toString(state.IsSealed)},
		{Key: labelPerfStandby, Value: toString(state.IsPerformanceStandby)},
		{Key: labelInitialized, Value: toString(state.IsInitialized)},
	}
	if err := client.UpdateServiceSelectors(namespace, serviceName, selectors...); err != nil {
		return nil, err
	}
	registration := &serviceRegistration{
		logger:      logger,
		serviceName: serviceName,
		namespace:   namespace,
		client:      client,
	}

	// Run a background goroutine to leave labels in the final state we'd like
	// when Vault shuts down.
	go registration.onShutdown(shutdownCh)
	return registration, nil
}

type serviceRegistration struct {
	logger                 log.Logger
	namespace, serviceName string
	client                 kubeHlpr.LightWeightClient
}

func (r *serviceRegistration) NotifyActiveStateChange(isActive bool) error {
	return r.client.UpdateServiceSelectors(r.namespace, r.serviceName, &kubeHlpr.Selector{
		Key:   labelActive,
		Value: toString(isActive),
	})
}

func (r *serviceRegistration) NotifySealedStateChange(isSealed bool) error {
	return r.client.UpdateServiceSelectors(r.namespace, r.serviceName, &kubeHlpr.Selector{
		Key:   labelSealed,
		Value: toString(isSealed),
	})
}

func (r *serviceRegistration) NotifyPerformanceStandbyStateChange(isStandby bool) error {
	return r.client.UpdateServiceSelectors(r.namespace, r.serviceName, &kubeHlpr.Selector{
		Key:   labelPerfStandby,
		Value: toString(isStandby),
	})
}

func (r *serviceRegistration) NotifyInitializedStateChange(isInitialized bool) error {
	return r.client.UpdateServiceSelectors(r.namespace, r.serviceName, &kubeHlpr.Selector{
		Key:   labelInitialized,
		Value: toString(isInitialized),
	})
}

func (r *serviceRegistration) onShutdown(shutdownCh <-chan struct{}) {
	<-shutdownCh

	// Label the service with the values we want to leave behind after shutdown.
	selectors := []*kubeHlpr.Selector{
		{Key: labelActive, Value: toString(false)},
		{Key: labelSealed, Value: toString(true)},
		{Key: labelPerfStandby, Value: toString(false)},
		{Key: labelInitialized, Value: toString(false)},
	}
	if err := r.client.UpdateServiceSelectors(r.namespace, r.serviceName, selectors...); err != nil {
		if r.logger.IsWarn() {
			r.logger.Warn(fmt.Sprintf("unable to set final status on service name %q in namespace %q on shutdown: %s", r.serviceName, r.namespace, err))
		}
		return
	}
}

// Converts a bool to "true" or "false".
func toString(b bool) string {
	return fmt.Sprintf("%t", b)
}
