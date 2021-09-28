package capability

import (
	"fmt"

	dynatracev1 "github.com/Dynatrace/dynatrace-operator/api/v1"
	"github.com/Dynatrace/dynatrace-operator/controllers/customproperties"
	"github.com/Dynatrace/dynatrace-operator/deploymentmetadata"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"path/filepath"
)

const (
	trustStoreVolume          = "truststore-volume"
	k8scrt2jksPath            = "/opt/dynatrace/gateway/k8scrt2jks.sh"
	activeGateCacertsPath     = "/opt/dynatrace/gateway/jre/lib/security/cacerts"
	activeGateSslPath         = "/var/lib/dynatrace/gateway/ssl"
	k8sCertificateFile        = "k8s-local.jks"
	k8scrt2jksWorkingDir      = "/var/lib/dynatrace/gateway"
	initContainerTemplateName = "certificate-loader"

	jettyCerts = "server-certs"

	secretsRootDir = "/var/lib/dynatrace/secrets/"

	serviceAccountPrefix = "dynatrace-"

	AnnotationVersion         = "internal.operator.dynatrace.com/version"
	AnnotationCustomPropsHash = "internal.operator.dynatrace.com/custom-properties-hash"

	DTCapabilities       = "DT_CAPABILITIES"
	DTIdSeedNamespace    = "DT_ID_SEED_NAMESPACE"
	DTIdSeedClusterId    = "DT_ID_SEED_K8S_CLUSTER_ID"
	DTNetworkZone        = "DT_NETWORK_ZONE"
	DTGroup              = "DT_GROUP"
	DTInternalProxy      = "DT_INTERNAL_PROXY"
	DTDeploymentMetadata = "DT_DEPLOYMENT_METADATA"

	ProxySecretKey = "proxy"
)

type MultiCapability struct {
	capabilityBase
}

type baseFunc func(*dynatracev1.DynaKube) *capabilityBase

var activeGateCapabilities = map[dynatracev1.ActiveGateCapability]baseFunc{
	dynatracev1.KubeMon:    kubeMonBase,
	dynatracev1.Routing:    routingBase,
	dynatracev1.DataIngest: dataIngestBase,
}

var ports = map[string]int{
	"kubemon":    9999,
	"routing":    9990,
	"data-ingest": 9980,
}


type Configuration struct {
	SetDnsEntryPoint     bool
	SetReadinessPort     bool
	SetCommunicationPort bool
	CreateService        bool
	ServiceAccountOwner  string
}

type Capability interface {
	Enabled() bool
	GetModuleName() string
	GetCapabilityName() string
	GetProperties() *dynatracev1.CapabilityProperties
	GetConfiguration() Configuration
	GetContainersTemplates() []corev1.Container
	GetInitContainersTemplates() []corev1.Container
	GetContainerVolumeMounts() []corev1.VolumeMount
	GetVolumes() []corev1.Volume
	// Deprecated
	GetServiceAccountName() string
}

type capabilityBase struct {
	enabled bool
	moduleName     string
	capabilityName string
	properties     *dynatracev1.CapabilityProperties
	Configuration
	containersTemplates     []corev1.Container
	initContainersTemplates []corev1.Container
	containerVolumeMounts   []corev1.VolumeMount
	volumes                 []corev1.Volume
	// Deprecated
	serviceAccountName string
}

func (c *capabilityBase) Enabled() bool {
	return c.enabled
}

func (c *capabilityBase) GetProperties() *dynatracev1.CapabilityProperties {
	return c.properties
}

func (c *capabilityBase) GetConfiguration() Configuration {
	return c.Configuration
}

func (c *capabilityBase) GetModuleName() string {
	return c.moduleName
}

func (c *capabilityBase) GetCapabilityName() string {
	return c.capabilityName
}

func (c *capabilityBase) GetContainersTemplates() []corev1.Container {
	return c.containersTemplates
}

// Note:
// Caller must set following fields:
//   Image:
//   Resources:
func (c *capabilityBase) GetInitContainersTemplates() []corev1.Container {
	return c.initContainersTemplates
}

func (c *capabilityBase) GetContainerVolumeMounts() []corev1.VolumeMount {
	return c.containerVolumeMounts
}

func (c *capabilityBase) GetVolumes() []corev1.Volume {
	return c.volumes
}

func (c *capabilityBase) GetServiceAccountName() string {
	return c.serviceAccountName
}

func CalculateStatefulSetName(capability Capability, instanceName string) string {
	return instanceName + "-" + capability.GetModuleName()
}

func (c *capabilityBase) setTlsVolumeMount(agSpec *dynatracev1.ActiveGateSpec) {
	if agSpec == nil {
		return
	}

	if agSpec.TlsSecretName != "" {
		c.containerVolumeMounts = append(c.containerVolumeMounts,
			corev1.VolumeMount{
				ReadOnly:  true,
				Name:      jettyCerts,
				MountPath: filepath.Join(secretsRootDir, "tls"),
			})
	}
}

func (c *capabilityBase) setTlsVolume(agSpec *dynatracev1.ActiveGateSpec) {
	if agSpec == nil {
		return
	}

	if agSpec.TlsSecretName != "" {
		c.volumes = append(c.volumes,
			corev1.Volume{
				Name: jettyCerts,
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName: agSpec.TlsSecretName,
					},
				},
			})
	}
}

func (cap *capabilityBase) buildInitContainers(dk *dynatracev1.DynaKube) []corev1.Container {
	ics := cap.initContainersTemplates

	for idx := range ics {
		ics[idx].Image = dk.ActiveGateImage()
		ics[idx].Resources = cap.properties.Resources
	}

	return ics
}

func (cap *capabilityBase) buildContainer(dk *dynatracev1.DynaKube) corev1.Container {
	return corev1.Container{
		Name:            cap.moduleName,
		Image:           dk.ActiveGateImage(),
		Resources:       cap.properties.Resources,
		ImagePullPolicy: corev1.PullAlways,
		Env:             cap.buildEnvs(dk),
		VolumeMounts:    cap.buildVolumeMounts(),
		ReadinessProbe: &corev1.Probe{
			Handler: corev1.Handler{
				HTTPGet: &corev1.HTTPGetAction{
					Path:   "/rest/health",
					Port:   intstr.IntOrString{IntVal: int32(ports[cap.moduleName])},
					Scheme: "HTTPS",
				},
			},
			InitialDelaySeconds: 90,
			PeriodSeconds:       15,
			FailureThreshold:    3,
		},
	}
}

func (cap *capabilityBase) buildVolumes(dk *dynatracev1.DynaKube) []corev1.Volume {
	var volumes []corev1.Volume

	if !isCustomPropertiesNilOrEmpty(cap.properties.CustomProperties) {
		valueFrom := cap.determineCustomPropertiesSource(dk)
		volumes = append(volumes, corev1.Volume{
			Name: customproperties.VolumeName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: valueFrom,
					Items: []corev1.KeyToPath{
						{Key: customproperties.DataKey, Path: customproperties.DataPath},
					},
				},
			},
		},
		)
	}

	volumes = append(volumes, cap.volumes...)

	return volumes
}

func (cap *capabilityBase) determineCustomPropertiesSource(dk *dynatracev1.DynaKube) string {
	if cap.properties.CustomProperties.ValueFrom == "" {
		return fmt.Sprintf("%s-%s-%s", dk.Name, cap.ServiceAccountOwner, customproperties.Suffix)
	}
	return cap.properties.CustomProperties.ValueFrom
}

func (cap *capabilityBase) buildVolumeMounts() []corev1.VolumeMount {
	var volumeMounts []corev1.VolumeMount

	if !isCustomPropertiesNilOrEmpty(cap.properties.CustomProperties) {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			ReadOnly:  true,
			Name:      customproperties.VolumeName,
			MountPath: customproperties.MountPath,
			SubPath:   customproperties.DataPath,
		})
	}

	volumeMounts = append(volumeMounts, cap.containerVolumeMounts...)

	return volumeMounts
}

func (cap *capabilityBase) buildEnvs(dk *dynatracev1.DynaKube) []corev1.EnvVar {
	deploymentMetadata := deploymentmetadata.NewDeploymentMetadata(string(dk.Status.KubeSystemUUID), deploymentmetadata.DeploymentTypeAG)

	envs := []corev1.EnvVar{
		{Name: DTCapabilities, Value: cap.capabilityName},
		{Name: DTIdSeedNamespace, Value: dk.Namespace},
		{Name: DTIdSeedClusterId, Value: string(dk.Status.KubeSystemUUID)},
		{Name: DTDeploymentMetadata, Value: deploymentMetadata.AsString()},
	}
	envs = append(envs, cap.properties.Env...)

	if !isProxyNilOrEmpty(dk.Spec.Proxy) {
		envs = append(envs, buildProxyEnv(dk.Spec.Proxy))
	}
	if cap.properties.Group != "" {
		envs = append(envs, corev1.EnvVar{Name: DTGroup, Value: cap.properties.Group})
	}
	if dk.Spec.NetworkZone != "" {
		envs = append(envs, corev1.EnvVar{Name: DTNetworkZone, Value: dk.Spec.NetworkZone})
	}

	return envs
}

func buildProxyEnv(proxy *dynatracev1.DynaKubeProxy) corev1.EnvVar {
	if proxy.ValueFrom != "" {
		return corev1.EnvVar{
			Name: DTInternalProxy,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: proxy.ValueFrom},
					Key:                  ProxySecretKey,
				},
			},
		}
	} else {
		return corev1.EnvVar{
			Name:  DTInternalProxy,
			Value: proxy.Value,
		}
	}
}

func isCustomPropertiesNilOrEmpty(customProperties *dynatracev1.DynaKubeValueSource) bool {
	return customProperties == nil ||
		(customProperties.Value == "" &&
			customProperties.ValueFrom == "")
}

func isProxyNilOrEmpty(proxy *dynatracev1.DynaKubeProxy) bool {
	return proxy == nil || (proxy.Value == "" && proxy.ValueFrom == "")
}


func NewMultiCapability(dk *dynatracev1.DynaKube) *MultiCapability {
	mc := MultiCapability{
		capabilityBase{
			moduleName: "multi",
			properties: &dk.Spec.ActiveGate.CapabilityProperties,
		},
	}
	if !dk.ActiveGateMode() || len(dk.Spec.ActiveGate.Capabilities) == 0{
		return &mc
	}
	mc.enabled = true
	for _, capName := range dk.Spec.ActiveGate.Capabilities {
		cap := activeGateCapabilities[capName](dk)
		if !mc.CreateService {
			mc.CreateService = cap.CreateService
		}
		if !mc.SetCommunicationPort {
			mc.SetCommunicationPort = cap.SetCommunicationPort
		}
		if !mc.SetDnsEntryPoint {
			mc.SetDnsEntryPoint = cap.SetDnsEntryPoint
		}
		if !mc.SetReadinessPort {
			mc.SetReadinessPort = cap.SetReadinessPort
		}
		if mc.ServiceAccountOwner != "" {
			mc.ServiceAccountOwner = cap.ServiceAccountOwner
		}

		cap.properties = mc.properties
		cap.containersTemplates = []corev1.Container{cap.buildContainer(dk)}
		cap.initContainersTemplates = cap.buildInitContainers(dk)
		cap.setTlsVolumeMount(&dk.Spec.ActiveGate)

		mc.containersTemplates = append(mc.containersTemplates, cap.containersTemplates...)
		mc.initContainersTemplates = append(mc.initContainersTemplates, cap.initContainersTemplates...)
		mc.containerVolumeMounts = append(mc.containerVolumeMounts, cap.containerVolumeMounts...)
		mc.volumes = append(mc.volumes, cap.volumes...)
	}
	mc.volumes = mc.buildVolumes(dk)
	mc.setTlsVolume(&dk.Spec.ActiveGate)
	return &mc

}

func kubeMonBase(dk *dynatracev1.DynaKube) *capabilityBase {
	c := capabilityBase{
		moduleName:     "kubemon",
		capabilityName: "kubernetes_monitoring",
		Configuration: Configuration{
			ServiceAccountOwner: "kubernetes-monitoring",
		},
		initContainersTemplates: []corev1.Container{
			{
				Name:            initContainerTemplateName,
				ImagePullPolicy: corev1.PullAlways,
				WorkingDir:      k8scrt2jksWorkingDir,
				Command:         []string{"/bin/bash"},
				Args:            []string{"-c", k8scrt2jksPath},
				VolumeMounts: []corev1.VolumeMount{
					{
						ReadOnly:  false,
						Name:      trustStoreVolume,
						MountPath: activeGateSslPath,
					},
				},
			},
		},
		containerVolumeMounts: []corev1.VolumeMount{{
			ReadOnly:  true,
			Name:      trustStoreVolume,
			MountPath: activeGateCacertsPath,
			SubPath:   k8sCertificateFile,
		}},
		volumes: []corev1.Volume{{
			Name: trustStoreVolume,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			}},
		},
	}
	return &c
}

func routingBase(dk *dynatracev1.DynaKube) *capabilityBase {
	c := capabilityBase{
		moduleName:     "routing",
		capabilityName: "MSGrouter",
		Configuration: Configuration{
			SetDnsEntryPoint:     true,
			SetReadinessPort:     true,
			SetCommunicationPort: true,
			CreateService:        true,
		},
	}
	return &c
}

func dataIngestBase(dk *dynatracev1.DynaKube) *capabilityBase {
	c := capabilityBase{
		moduleName:     "data-ingest",
		capabilityName: "metrics_ingest",
		Configuration: Configuration{
			SetDnsEntryPoint:     true,
			SetReadinessPort:     true,
			SetCommunicationPort: true,
			CreateService:        true,
		},
	}
	return &c
}