package main

import (
	"fmt"
	"os"

	"github.com/Azure/go-autorest/autorest/azure"
)

type AzureConfiguration struct {
	ClientID              string
	ClientSecret          string
	SubscriptionID        string
	TenantID              string
	KeyVaultKeyIdentifier string
}

const cloudName string = "AzurePublicCloud"

func Environment() *azure.Environment {
	env, err := azure.EnvironmentFromName(cloudName)
	if err != nil {
		panic(fmt.Sprintf(
			"invalid cloud name '%s' specified, cannot continue\n", cloudName))
	}
	return &env
}

func ParseEnvironment() (AzureConfiguration, error) {
	clientID, err := getMustEnv("AZURE_CLIENT_ID")
	if err != nil {
		return AzureConfiguration{}, err
	}

	clientSecret, err := getMustEnv("AZURE_CLIENT_SECRET")
	if err != nil {
		return AzureConfiguration{}, err
	}

	subscriptionID, err := getMustEnv("AZURE_SUBSCRIPTION_ID")
	if err != nil {
		return AzureConfiguration{}, err
	}

	tenantID, err := getMustEnv("AZURE_TENANT_ID")
	if err != nil {
		return AzureConfiguration{}, err
	}

	KeyVaultKeyIdentifier, err := getMustEnv("AZURE_KEY_VAULT_KEY_IDENTIFIER")
	if err != nil {
		return AzureConfiguration{}, err
	}

	return AzureConfiguration{clientID, clientSecret, subscriptionID, tenantID, KeyVaultKeyIdentifier}, nil
}

func getMustEnv(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", fmt.Errorf("expected env vars not provided: %s", key)
	}
	return value, nil
}
