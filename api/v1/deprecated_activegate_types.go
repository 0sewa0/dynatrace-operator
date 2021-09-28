package v1

// Deprecated
type RoutingSpec struct {
	// Enables Capability
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Capability",order=29,xDescriptors="urn:alm:descriptor:com.tectonic.ui:selector:booleanSwitch"
	Enabled bool `json:"enabled,omitempty"`

	CapabilityProperties `json:",inline"`

	// Optional: set custom Service Account Name used with ActiveGate pods
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Service Account name",order=40,xDescriptors={"urn:alm:descriptor:com.tectonic.ui:advanced","urn:alm:descriptor:io.kubernetes:ServiceAccount"}
	ServiceAccountName string `json:"serviceAccountName,omitempty"`
}

// Deprecated
type KubernetesMonitoringSpec struct {
	// Enables Capability
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Capability",order=29,xDescriptors="urn:alm:descriptor:com.tectonic.ui:selector:booleanSwitch"
	Enabled bool `json:"enabled,omitempty"`

	CapabilityProperties `json:",inline"`

	// Optional: set custom Service Account Name used with ActiveGate pods
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Service Account name",order=40,xDescriptors={"urn:alm:descriptor:com.tectonic.ui:advanced","urn:alm:descriptor:io.kubernetes:ServiceAccount"}
	ServiceAccountName string `json:"serviceAccountName,omitempty"`
}
