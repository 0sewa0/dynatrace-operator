package statefulset

import (
	dynatracev1 "github.com/Dynatrace/dynatrace-operator/api/v1"
	"github.com/Dynatrace/dynatrace-operator/controllers/activegate/capability"
	"github.com/Dynatrace/dynatrace-operator/controllers/activegate/internal/events"
	"github.com/Dynatrace/dynatrace-operator/controllers/kubeobjects"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
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

type statefulSetProperties struct {
	dk                     dynatracev1.DynaKube
	capability             capability.Capability
	customPropertiesHash   string
	majorKubernetesVersion string
	minorKubernetesVersion string
	OnAfterCreateListener  []events.StatefulSetEvent
}

func NewStatefulSetProperties(dk dynatracev1.DynaKube, capability capability.Capability, customPropertiesHash string, majorKubernetesVersion string, minorKubernetesVersion string) *statefulSetProperties {

	return &statefulSetProperties{
		dk:                     dk,
		capability:             capability,
		customPropertiesHash:   customPropertiesHash,
		majorKubernetesVersion: majorKubernetesVersion,
		minorKubernetesVersion: minorKubernetesVersion,
		OnAfterCreateListener:  []events.StatefulSetEvent{},
	}
}

func (stsProperties *statefulSetProperties) CreateStatefulSet() (*appsv1.StatefulSet, error) {
	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:        stsProperties.dk.Name + "-" + stsProperties.capability.GetModuleName(),
			Namespace:   stsProperties.dk.Namespace,
			Labels:      buildLabels(&stsProperties.dk, stsProperties.capability.GetModuleName(), stsProperties.capability.GetProperties()),
			Annotations: map[string]string{},
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas:            stsProperties.capability.GetProperties().Replicas,
			PodManagementPolicy: appsv1.ParallelPodManagement,
			Selector:            &metav1.LabelSelector{MatchLabels: BuildLabelsFromInstance(&stsProperties.dk, stsProperties.capability.GetModuleName())},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: buildLabels(&stsProperties.dk, stsProperties.capability.GetModuleName(), stsProperties.capability.GetProperties()),
					Annotations: map[string]string{
						AnnotationVersion:         stsProperties.dk.Status.ActiveGate.Version,
						AnnotationCustomPropsHash: stsProperties.customPropertiesHash,
					},
				},
				Spec: stsProperties.buildTemplateSpec(),
			},
		}}

	for _, onAfterCreateListener := range stsProperties.OnAfterCreateListener {
		onAfterCreateListener(sts)
	}

	hash, err := kubeobjects.GenerateHash(sts)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	sts.ObjectMeta.Annotations[kubeobjects.AnnotationHash] = hash
	return sts, nil
}

func (stsProperties *statefulSetProperties) buildTemplateSpec() corev1.PodSpec {
	return corev1.PodSpec{
		Containers:         stsProperties.capability.GetContainersTemplates(),
		InitContainers:     stsProperties.capability.GetInitContainersTemplates(),
		NodeSelector:       stsProperties.capability.GetProperties().NodeSelector,
		ServiceAccountName: stsProperties.capability.GetServiceAccountName(),
		Affinity:           affinity(stsProperties),
		Tolerations:        stsProperties.capability.GetProperties().Tolerations,
		Volumes:            stsProperties.capability.GetVolumes(),
		ImagePullSecrets: []corev1.LocalObjectReference{
			{Name: stsProperties.dk.PullSecret()},
		},
	}
}
