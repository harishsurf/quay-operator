package validation

import (
	"context"
	"fmt"
	"reflect"

	"time"

	"github.com/redhat-cop/quay-operator/pkg/controller/quayecosystem/constants"
	"github.com/redhat-cop/quay-operator/pkg/controller/quayecosystem/logging"
	"github.com/redhat-cop/quay-operator/pkg/controller/quayecosystem/resources"
	"github.com/redhat-cop/quay-operator/pkg/controller/quayecosystem/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Validate performs validation across all resources
func Validate(client client.Client, quayConfiguration *resources.QuayConfiguration) (bool, error) {

	// Validate Superuser Credentials Secret
	if !utils.IsZeroOfUnderlyingType(quayConfiguration.QuayEcosystem.Spec.Quay.SuperuserCredentialsSecretName) {

		validQuaySuperuserSecret, superuserSecret, err := validateSecret(client, quayConfiguration.QuayEcosystem.Namespace, quayConfiguration.QuayEcosystem.Spec.Quay.SuperuserCredentialsSecretName, constants.DefaultQuaySuperuserCredentials)

		if err != nil {
			return false, err
		}

		if !validQuaySuperuserSecret {
			return false, fmt.Errorf("Failed to validate provided Quay Superuser Secret")
		}

		quayConfiguration.QuaySuperuserEmail = string(superuserSecret.Data[constants.QuaySuperuserEmailKey])
		quayConfiguration.QuaySuperuserUsername = string(superuserSecret.Data[constants.QuaySuperuserUsernameKey])
		quayConfiguration.QuaySuperuserPassword = string(superuserSecret.Data[constants.QuaySuperuserPasswordKey])
		quayConfiguration.ValidProvidedQuaySuperuserSecret = true
	}

	if len(quayConfiguration.QuaySuperuserPassword) < 8 {
		return false, fmt.Errorf("Quay Superuser Password Must Be At Least 8 Characters in Length")
	}

	if !utils.IsZeroOfUnderlyingType(quayConfiguration.QuayEcosystem.Spec.Quay.ConfigSecretName) {

		validQuayConfigSecret, quayConfigSecret, err := validateSecret(client, quayConfiguration.QuayEcosystem.Namespace, quayConfiguration.QuayEcosystem.Spec.Quay.ConfigSecretName, constants.DefaultQuayConfigCredentials)

		if err != nil {
			return false, err
		}

		if !validQuayConfigSecret {
			return false, fmt.Errorf("Failed to validate provided Quay Config Secret")
		}

		quayConfiguration.QuayConfigPassword = string(quayConfigSecret.Data[constants.QuayConfigPasswordKey])
		quayConfiguration.QuayConfigPasswordSecret = quayConfiguration.QuayEcosystem.Spec.Quay.ConfigSecretName
		quayConfiguration.ValidProvidedQuayConfigPasswordSecret = true

	}

	// Validate Quay ImagePullSecret
	if !utils.IsZeroOfUnderlyingType(quayConfiguration.QuayEcosystem.Spec.Quay.ImagePullSecretName) {

		validImagePullSecret, _, err := validateSecret(client, quayConfiguration.QuayEcosystem.Namespace, quayConfiguration.QuayEcosystem.Spec.Quay.ImagePullSecretName, nil)

		if err != nil {
			return false, err
		}

		if !validImagePullSecret {
			return false, fmt.Errorf("Failed to validate provided Quay Image Pull Secret")
		}

	}

	// Validate Redis ImagePullSecret
	if !utils.IsZeroOfUnderlyingType(quayConfiguration.QuayEcosystem.Spec.Redis.ImagePullSecretName) {

		validImagePullSecret, _, err := validateSecret(client, quayConfiguration.QuayEcosystem.Namespace, quayConfiguration.QuayEcosystem.Spec.Redis.ImagePullSecretName, nil)

		if err != nil {
			return false, err
		}

		if !validImagePullSecret {
			return false, fmt.Errorf("Failed to validate provided Redis Image Pull Secret")
		}
	}

	// Validate Redis CredentialsSecret
	if !utils.IsZeroOfUnderlyingType(quayConfiguration.QuayEcosystem.Spec.Redis.CredentialsSecretName) {

		validRedisCredentialSecret, redisSecret, err := validateSecret(client, quayConfiguration.QuayEcosystem.Namespace, quayConfiguration.QuayEcosystem.Spec.Redis.CredentialsSecretName, []string{constants.RedisPasswordKey})

		if err != nil {
			return false, err
		}

		if !validRedisCredentialSecret {
			return false, fmt.Errorf("Failed to validate provided Redis Credentials Secret")
		}

		quayConfiguration.RedisPassword = string(redisSecret.Data[constants.RedisPasswordKey])
		quayConfiguration.ValidProvidedRedisPasswordSecret = true
	}

	// Validate Quay Database ImagePullSecret
	if !utils.IsZeroOfUnderlyingType(quayConfiguration.QuayEcosystem.Spec.Quay.Database.ImagePullSecretName) {

		validImagePullSecret, _, err := validateSecret(client, quayConfiguration.QuayEcosystem.Namespace, quayConfiguration.QuayEcosystem.Spec.Quay.Database.ImagePullSecretName, nil)

		if err != nil {
			return false, err
		}

		if !validImagePullSecret {
			return false, fmt.Errorf("Failed to validate provided Data Database Image Pull Secret")
		}
	}

	// Validate Quay Database Credential
	if !quayConfiguration.QuayEcosystem.Spec.Quay.SkipSetup {
		if !utils.IsZeroOfUnderlyingType(quayConfiguration.QuayEcosystem.Spec.Quay.Database) && !utils.IsZeroOfUnderlyingType(quayConfiguration.QuayEcosystem.Spec.Quay.Database.Server) && utils.IsZeroOfUnderlyingType(quayConfiguration.QuayEcosystem.Spec.Quay.Database.CredentialsSecretName) {
			return false, fmt.Errorf("Failed to locate a Quay Database Credential for Externally Provisioned Instance")
		}

		if !utils.IsZeroOfUnderlyingType(quayConfiguration.QuayEcosystem.Spec.Quay.Database) && !utils.IsZeroOfUnderlyingType(quayConfiguration.QuayEcosystem.Spec.Quay.Database.CredentialsSecretName) {

			validQuayDatabaseSecret, databaseSecret, err := validateSecret(client, quayConfiguration.QuayEcosystem.Namespace, quayConfiguration.QuayEcosystem.Spec.Quay.Database.CredentialsSecretName, constants.RequiredDatabaseCredentialKeys)

			if err != nil {
				return false, err
			}

			if !validQuayDatabaseSecret {
				return false, fmt.Errorf("Failed to validate provided Quay Database Secret")
			}

			quayConfiguration.QuayDatabase.Username = string(databaseSecret.Data[constants.DatabaseCredentialsUsernameKey])
			quayConfiguration.QuayDatabase.Password = string(databaseSecret.Data[constants.DatabaseCredentialsPasswordKey])
			quayConfiguration.QuayDatabase.Database = string(databaseSecret.Data[constants.DatabaseCredentialsDatabaseKey])

			if _, found := databaseSecret.Data[constants.DatabaseCredentialsRootPasswordKey]; found {
				quayConfiguration.QuayDatabase.RootPassword = string(databaseSecret.Data[constants.DatabaseCredentialsRootPasswordKey])
			}

			quayConfiguration.ValidProvidedQuayDatabaseSecret = true
		}
	}

	// Validate Quay Database
	if !utils.IsZeroOfUnderlyingType(quayConfiguration.QuayEcosystem.Spec.Quay.Database.VolumeSize) {

		_, err := resource.ParseQuantity(quayConfiguration.QuayEcosystem.Spec.Quay.Database.VolumeSize)

		if err != nil {
			return false, err
		}
	}

	// Quay PVC Generation
	if !utils.IsZeroOfUnderlyingType(quayConfiguration.QuayEcosystem.Spec.Quay.RegistryStorage) {

		_, err := resource.ParseQuantity(quayConfiguration.QuayEcosystem.Spec.Quay.RegistryStorage.PersistentVolumeSize)

		if err != nil {
			return false, err
		}

	}

	// Registry Backends
	for _, registryBackend := range quayConfiguration.QuayEcosystem.Spec.Quay.RegistryBackends {

		// Validate replication is not enabled when using a Local backend
		if quayConfiguration.QuayEcosystem.Spec.Quay.EnableStorageReplication {
			if registryBackend.Local != nil {
				return false, fmt.Errorf("Cannot have make use of local storage when replication enabled. Local storage: %s", registryBackend.Name)
			}
		}

		managedRegistryBackend := registryBackend.DeepCopy()

		// Validate various backends
		if !utils.IsZeroOfUnderlyingType(managedRegistryBackend.S3) {

			// TODO: Do basic field validation
			if !utils.IsZeroOfUnderlyingType(managedRegistryBackend.S3.CredentialsSecretName) {
				validS3Secret, s3Secret, err := validateSecret(client, quayConfiguration.QuayEcosystem.Namespace, registryBackend.S3.CredentialsSecretName, constants.RequiredS3CredentialKeys)

				if err != nil {
					return false, err
				}
				if !validS3Secret {
					return false, fmt.Errorf("Failed to validate provided registry backend. Name: %s", managedRegistryBackend.Name)
				}

				managedRegistryBackend.S3.AccessKey = string(s3Secret.Data[constants.S3AccessKey])
				managedRegistryBackend.S3.SecretKey = string(s3Secret.Data[constants.S3SecretKey])
				managedRegistryBackend.S3.CredentialsSecretName = ""

			}

		}

		quayConfiguration.RegistryBackends = append(quayConfiguration.RegistryBackends, *managedRegistryBackend)

	}

	// Validate Quay SSL Certificates
	if !utils.IsZeroOfUnderlyingType(quayConfiguration.QuayEcosystem.Spec.Quay.SslCertificatesSecretName) {
		validQuaySslCertificateSecret, quaySslCertificateSecret, err := validateSecret(client, quayConfiguration.QuayEcosystem.Namespace, quayConfiguration.QuayEcosystem.Spec.Quay.SslCertificatesSecretName, constants.RequiredSslCertificateKeys)

		if err != nil {
			return false, err
		}

		if !validQuaySslCertificateSecret {
			return false, fmt.Errorf("Failed to validate provided Quay SSL Certificate")
		}

		quayConfiguration.QuaySslCertificate = quaySslCertificateSecret.Data[constants.QuayAppConfigSSLCertificateSecretKey]
		quayConfiguration.QuaySslPrivateKey = quaySslCertificateSecret.Data[constants.QuayAppConfigSSLPrivateKeySecretKey]

	}

	if quayConfiguration.QuayEcosystem.Spec.Clair != nil && quayConfiguration.QuayEcosystem.Spec.Clair.Enabled {

		// Validate Clair ImagePullSecret
		if !utils.IsZeroOfUnderlyingType(quayConfiguration.QuayEcosystem.Spec.Clair.ImagePullSecretName) {

			validImagePullSecret, _, err := validateSecret(client, quayConfiguration.QuayEcosystem.Namespace, quayConfiguration.QuayEcosystem.Spec.Clair.ImagePullSecretName, nil)

			if err != nil {
				return false, err
			}

			if !validImagePullSecret {
				return false, fmt.Errorf("Failed to validate provided Clair Image Pull Secret")
			}

		}

		// Validate Update Interval
		if !utils.IsZeroOfUnderlyingType(quayConfiguration.QuayEcosystem.Spec.Clair.UpdateInterval) {

			duration, durationErr := time.ParseDuration(quayConfiguration.QuayEcosystem.Spec.Clair.UpdateInterval)

			if durationErr != nil {
				return false, durationErr
			}

			quayConfiguration.ClairUpdateInterval = duration
		}

		if !utils.IsZeroOfUnderlyingType(quayConfiguration.QuayEcosystem.Spec.Clair.Database) && !utils.IsZeroOfUnderlyingType(quayConfiguration.QuayEcosystem.Spec.Clair.Database.Server) && utils.IsZeroOfUnderlyingType(quayConfiguration.QuayEcosystem.Spec.Clair.Database.CredentialsSecretName) {
			return false, fmt.Errorf("Failed to locate a Clair Database Credential for Externally Provisioned Instance")
		}

		if !utils.IsZeroOfUnderlyingType(quayConfiguration.QuayEcosystem.Spec.Clair.Database) && !utils.IsZeroOfUnderlyingType(quayConfiguration.QuayEcosystem.Spec.Clair.Database.CredentialsSecretName) {

			validClairDatabaseSecret, databaseSecret, err := validateSecret(client, quayConfiguration.QuayEcosystem.Namespace, quayConfiguration.QuayEcosystem.Spec.Clair.Database.CredentialsSecretName, constants.RequiredDatabaseCredentialKeys)

			if err != nil {
				return false, err
			}

			if !validClairDatabaseSecret {
				return false, fmt.Errorf("Failed to validate provided Clair Database Secret")
			}

			quayConfiguration.ClairDatabase.Username = string(databaseSecret.Data[constants.DatabaseCredentialsUsernameKey])
			quayConfiguration.ClairDatabase.Password = string(databaseSecret.Data[constants.DatabaseCredentialsPasswordKey])
			quayConfiguration.ClairDatabase.Database = string(databaseSecret.Data[constants.DatabaseCredentialsDatabaseKey])

			if _, found := databaseSecret.Data[constants.DatabaseCredentialsRootPasswordKey]; found {
				quayConfiguration.ClairDatabase.RootPassword = string(databaseSecret.Data[constants.DatabaseCredentialsRootPasswordKey])
			}

			quayConfiguration.ValidProvidedClairDatabaseSecret = true
		}

	}

	return true, nil
}

func validateSecret(client client.Client, namespace string, name string, requiredParameters interface{}) (bool, *corev1.Secret, error) {

	secret := &corev1.Secret{}
	err := client.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: name}, secret)
	if err != nil && errors.IsNotFound(err) {
		logging.Log.Error(fmt.Errorf("Secret not Found"), "Secret Validation", "Namespace", namespace, "Name", name)
		return false, nil, err
	} else if err != nil && !errors.IsNotFound(err) {
		logging.Log.Error(fmt.Errorf("Error retrieving secret"), "Secret Validation", "Namespace", namespace, "Name", name)
		return false, nil, err
	}

	if requiredParameters != nil {

		validSecret := false
		if reflect.TypeOf(requiredParameters).Kind() == reflect.Map {
			validSecret = validateProvidedSecretMap(secret, requiredParameters.(map[string]string))

		}
		if reflect.TypeOf(requiredParameters).Kind() == reflect.Slice {
			validSecret = validateProvidedSecretSlice(secret, requiredParameters.([]string))

		}

		if !validSecret {
			logging.Log.Error(fmt.Errorf("Failed to validate provided secret with required parameters"), "Secret Validation", "Namespace", namespace, "Name", name)
			return false, secret, fmt.Errorf("Failed to validate provided secret with required parameters. Namespace: %s, Name: %s", namespace, name)
		}
	}

	return true, secret, nil

}

func validateProvidedSecretMap(secret *corev1.Secret, requiredParameters map[string]string) bool {

	for key := range requiredParameters {
		if _, found := secret.Data[key]; !found {
			return false
		}
	}

	return true

}

func validateProvidedSecretSlice(secret *corev1.Secret, requiredParameters []string) bool {

	for _, value := range requiredParameters {
		if _, found := secret.Data[value]; !found {
			return false
		}
	}

	return true

}
