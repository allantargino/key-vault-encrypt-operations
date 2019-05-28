package main

import (
	"fmt"
	"os"
)

type AzureConfiguration struct {
	ClientID              string
	ClientSecret          string
	TenantID              string
	KeyVaultKeyIdentifier string
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
	tenantID, err := getMustEnv("AZURE_TENANT_ID")
	if err != nil {
		return AzureConfiguration{}, err
	}

	KeyVaultKeyIdentifier, err := getMustEnv("AZURE_KEY_VAULT_KEY_IDENTIFIER")
	if err != nil {
		return AzureConfiguration{}, err
	}

	return AzureConfiguration{clientID, clientSecret, tenantID, KeyVaultKeyIdentifier}, nil
}

func getMustEnv(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", fmt.Errorf("expected env vars not provided: %s", key)
	}
	return value, nil
}
