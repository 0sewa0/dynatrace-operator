package capability

import (
	dynatracev1 "github.com/Dynatrace/dynatrace-operator/api/v1"

	corev1 "k8s.io/api/core/v1"
)

// Deprecated
type KubeMonCapability struct {
	capabilityBase
}

// Deprecated
type RoutingCapability struct {
	capabilityBase
}

// Deprecated
func NewKubeMonCapability(dk *dynatracev1.DynaKube) *KubeMonCapability {
	c := &KubeMonCapability{
		*kubeMonBase(dk),
	}
	if !dk.Spec.KubernetesMonitoring.Enabled {
		return c
	}
	c.enabled = true
	c.properties = &dk.Spec.KubernetesMonitoring.CapabilityProperties
	c.serviceAccountName = c.determineServiceAccountName()
	c.containersTemplates = []corev1.Container{c.buildContainer(dk)}
	c.initContainersTemplates = c.buildInitContainers(dk)
	return c
}

// Deprecated
func NewRoutingCapability(dk *dynatracev1.DynaKube) *RoutingCapability {
	c := &RoutingCapability{
		*routingBase(dk),
	}
	if !dk.Spec.Routing.Enabled {
		return c
	}
	c.enabled = true
	c.properties = &dk.Spec.Routing.CapabilityProperties
	c.serviceAccountName = c.determineServiceAccountName()
	c.initContainersTemplates = c.buildInitContainers(dk)
	return c
}

// Deprecated
func (cap *capabilityBase) determineServiceAccountName() string {
	if cap.serviceAccountName == "" {
		return serviceAccountPrefix + cap.ServiceAccountOwner
	}
	return cap.serviceAccountName
}