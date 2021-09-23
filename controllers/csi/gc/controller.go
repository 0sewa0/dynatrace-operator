package csigc

import (
	"context"
	"time"

	dynatracev1 "github.com/Dynatrace/dynatrace-operator/api/v1"
	dtcsi "github.com/Dynatrace/dynatrace-operator/controllers/csi"
	"github.com/Dynatrace/dynatrace-operator/controllers/csi/metadata"
	"github.com/Dynatrace/dynatrace-operator/controllers/dynakube"
	"github.com/Dynatrace/dynatrace-operator/dtclient"
	"github.com/go-logr/logr"
	"github.com/spf13/afero"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// CSIGarbageCollector removes unused and outdated agent versions
type CSIGarbageCollector struct {
	client       client.Client
	logger       logr.Logger
	opts         dtcsi.CSIOptions
	dtcBuildFunc dynakube.DynatraceClientFunc
	fs           afero.Fs
	db           metadata.Access
	path         metadata.PathResolver
}

// NewReconciler returns a new CSIGarbageCollector
func NewReconciler(client client.Client, opts dtcsi.CSIOptions, db metadata.Access) *CSIGarbageCollector {
	return &CSIGarbageCollector{
		client:       client,
		logger:       log.Log.WithName("csi.gc.controller"),
		opts:         opts,
		dtcBuildFunc: dynakube.BuildDynatraceClient,
		fs:           afero.NewOsFs(),
		db:           db,
		path:         metadata.PathResolver{RootDir: opts.RootDir},
	}
}

func (gc *CSIGarbageCollector) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dynatracev1.DynaKube{}).
		Complete(gc)
}

var _ reconcile.Reconciler = &CSIGarbageCollector{}

func (gc *CSIGarbageCollector) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	gc.logger.Info("running OneAgent garbage collection", "namespace", request.Namespace, "name", request.Name)
	reconcileResult := reconcile.Result{RequeueAfter: 60 * time.Minute}

	var dk dynatracev1.DynaKube
	if err := gc.client.Get(ctx, request.NamespacedName, &dk); err != nil {
		if k8serrors.IsNotFound(err) {
			gc.logger.Info("given DynaKube object not found")
			return reconcileResult, nil
		}

		gc.logger.Error(err, "failed to get DynaKube object")
		return reconcileResult, nil
	}

	var tokens corev1.Secret
	if err := gc.client.Get(ctx, client.ObjectKey{Name: dk.Tokens(), Namespace: dk.Namespace}, &tokens); err != nil {
		gc.logger.Error(err, "failed to query tokens")
		return reconcileResult, nil
	}

	dtp := dynakube.DynatraceClientProperties{
		Client:              gc.client,
		Secret:              &tokens,
		ApiUrl:              dk.Spec.APIURL,
		Proxy:               (*dynakube.DynatraceClientProxy)(dk.Spec.Proxy),
		Namespace:           dk.Namespace,
		NetworkZone:         dk.Spec.NetworkZone,
		TrustedCerts:        dk.Spec.TrustedCAs,
		SkipCertCheck:       dk.Spec.SkipCertCheck,
		DisableHostRequests: dk.FeatureDisableHostsRequests(),
	}
	dtc, err := gc.dtcBuildFunc(dtp)
	if err != nil {
		gc.logger.Error(err, "failed to create Dynatrace client")
		return reconcileResult, nil
	}

	ci, err := dtc.GetConnectionInfo()
	if err != nil {
		gc.logger.Info("failed to fetch connection info")
		return reconcileResult, nil
	}

	latestAgentVersion, err := dtc.GetLatestAgentVersion(dtclient.OsUnix, dtclient.InstallerTypePaaS)
	if err != nil {
		gc.logger.Info("failed to query OneAgent version")
		return reconcileResult, nil
	}

	gc.logger.Info("running binary garbage collection")
	gc.runBinaryGarbageCollection(ci.TenantUUID, latestAgentVersion)

	gc.logger.Info("running log garbage collection")
	gc.runLogGarbageCollection(ci.TenantUUID)

	return reconcileResult, nil
}
