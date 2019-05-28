package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/Azure/azure-sdk-for-go/services/keyvault/2016-10-01/keyvault"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
)

const kvClientUserAgent = "terraform-remote-state"

type EncryptionClient struct {
	kvClient   keyvault.BaseClient
	vaultURL   string
	keyName    string
	keyVersion string
}

func NewEncryptionClient(azureConfiguration AzureConfiguration) (*EncryptionClient, error) {
	authorizer, err := getKeyvaultAuthorizer(azureConfiguration)
	if err != nil {
		return &EncryptionClient{}, err
	}

	kvClient := getKeysClient(authorizer)

	r, _ := regexp.Compile("https://(.*)\\.vault\\.azure\\.net/keys/(.*)/(.*)")

	str := r.FindAllString(azureConfiguration.KeyVaultKeyIdentifier, -1)
	if len(str) < 3 {
		return &EncryptionClient{}, fmt.Errorf("Expected a key identifier from Key Vault. e.g.: https://keyvaultname.vault.azure.net/keys/myKey/99d67321dd9841af859129cd5551a871")
	}

	fmt.Println(str)

	return &EncryptionClient{kvClient, "", "", ""}, nil
}

func getKeyvaultAuthorizer(azureConfiguration AzureConfiguration) (autorest.Authorizer, error) {
	// BUG: default value for KeyVaultEndpoint is wrong
	vaultEndpoint := strings.TrimSuffix(Environment().KeyVaultEndpoint, "/")
	// BUG: alternateEndpoint replaces other endpoints in the configs below
	alternateEndpoint, _ := url.Parse(
		"https://login.windows.net/" + azureConfiguration.TenantID + "/oauth2/token")

	var a autorest.Authorizer
	var err error

	oauthconfig, err := adal.NewOAuthConfig(Environment().ActiveDirectoryEndpoint, azureConfiguration.TenantID)
	if err != nil {
		return a, err
	}
	oauthconfig.AuthorizeEndpoint = *alternateEndpoint

	token, err := adal.NewServicePrincipalToken(
		*oauthconfig, azureConfiguration.ClientID, azureConfiguration.ClientSecret, vaultEndpoint)

	if err != nil {
		return a, err
	}

	keyvaultAuthorizer := autorest.NewBearerAuthorizer(token)

	return keyvaultAuthorizer, err
}

func getKeysClient(authorizer autorest.Authorizer) keyvault.BaseClient {
	keyClient := keyvault.New()
	keyClient.Authorizer = authorizer
	keyClient.AddToUserAgent(kvClientUserAgent)
	return keyClient
}

func (e *EncryptionClient) getKeyOperationsParameters(value *string) keyvault.KeyOperationsParameters {
	parameters := keyvault.KeyOperationsParameters{}
	parameters.Algorithm = keyvault.RSAOAEP256
	parameters.Value = value
	return parameters
}

func (e *EncryptionClient) Encrypt(ctx context.Context, data []byte) (*string, error) {
	encoded := base64.RawStdEncoding.EncodeToString(data)
	parameters := e.getKeyOperationsParameters(&encoded)
	result, err := e.kvClient.Encrypt(ctx, e.vaultURL, e.keyName, e.keyVersion, parameters)
	if err != nil {
		return nil, err
	}
	return result.Result, nil
}

func (e *EncryptionClient) Decrypt(ctx context.Context, data *string) ([]byte, error) {
	parameters := e.getKeyOperationsParameters(data)
	result, err := e.kvClient.Decrypt(ctx, e.vaultURL, e.keyName, e.keyVersion, parameters)
	if err != nil {
		return nil, err
	}
	decoded, err := base64.RawStdEncoding.DecodeString(*result.Result)
	if err != nil {
		return nil, err
	}
	return decoded, nil
}
