package kubeobjects

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/Dynatrace/dynatrace-operator/dtclient"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// CreateOrUpdateSecretIfNotExists creates a secret in case it does not exist or updates it if there are changes
func CreateOrUpdateSecretIfNotExists(c client.Client, r client.Reader, secretName string, targetNS string, data map[string][]byte, secretType corev1.SecretType, log logr.Logger) error {
	var cfg corev1.Secret
	err := r.Get(context.TODO(), client.ObjectKey{Name: secretName, Namespace: targetNS}, &cfg)
	if k8serrors.IsNotFound(err) {
		log.Info("Creating OneAgent config secret")
		if err := c.Create(context.TODO(), &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: targetNS,
			},
			Type: secretType,
			Data: data,
		}); err != nil {
			return errors.Wrapf(err, "failed to create secret %s", secretName)
		}
		return nil
	}

	if err != nil {
		return errors.Wrapf(err, "failed to query for secret %s", secretName)
	}

	if !reflect.DeepEqual(data, cfg.Data) {
		log.Info(fmt.Sprintf("Updating secret %s", secretName))
		cfg.Data = data
		if err := c.Update(context.TODO(), &cfg); err != nil {
			return errors.Wrapf(err, "failed to update secret %s", secretName)
		}
	}

	return nil
}

type Tokens struct {
	ApiToken  string
	PaasToken string
}

func NewTokens(secret *corev1.Secret) (*Tokens, error) {
	if secret == nil {
		return nil, fmt.Errorf("could not parse tokens: secret is nil")
	}

	var apiToken string
	var paasToken string
	var err error

	if err = verifySecret(secret); err != nil {
		return nil, errors.WithStack(err)
	}

	//Errors would have been caught by verifySecret
	apiToken, _ = ExtractToken(secret, dtclient.DynatraceApiToken)
	paasToken, _ = ExtractToken(secret, dtclient.DynatracePaasToken)

	return &Tokens{
		ApiToken:  apiToken,
		PaasToken: paasToken,
	}, nil
}

func verifySecret(secret *corev1.Secret) error {
	for _, token := range []string{
		dtclient.DynatraceApiToken,
		dtclient.DynatracePaasToken} {
		_, err := ExtractToken(secret, token)
		if err != nil {
			return errors.Errorf("invalid secret %s, %s", secret.Name, err)
		}
	}

	return nil
}

func ExtractToken(secret *corev1.Secret, key string) (string, error) {
	value, hasKey := secret.Data[key]
	if !hasKey {
		err := fmt.Errorf("missing token %s", key)
		return "", err
	}

	return strings.TrimSpace(string(value)), nil
}
