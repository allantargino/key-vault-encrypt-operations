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
	"github.com/Azure/go-autorest/autorest/azure"
)

const (
	kvClientUserAgent = "terraform-remote-state"
	cloudName         = "AzurePublicCloud"
)

type KeyVaultKeyInfo struct {
	vaultURL   string
	keyName    string
	keyVersion string
}

type EncryptionClient struct {
	kvClient *keyvault.BaseClient
	kvInfo   *KeyVaultKeyInfo
}

func NewEncryptionClientFromEnv(azureConfiguration AzureConfiguration) (*EncryptionClient, error) {
	return NewEncryptionClient(azureConfiguration.TenantID, azureConfiguration.ClientID, azureConfiguration.ClientSecret, azureConfiguration.KeyVaultKeyIdentifier)
}

func NewEncryptionClient(tenantID, clientID, clientSecret, keyVaultKeyIdentifier string) (*EncryptionClient, error) {
	authorizer, err := getKeyvaultAuthorizer(tenantID, clientID, clientSecret)
	if err != nil {
		return &EncryptionClient{}, err
	}

	kvClient := getKeysClient(authorizer)

	kvInfo, err := parseKeyVaultKeyInfo(keyVaultKeyIdentifier)
	if err != nil {
		return &EncryptionClient{}, err
	}

	return &EncryptionClient{kvClient, kvInfo}, nil
}

func environment() *azure.Environment {
	env, err := azure.EnvironmentFromName(cloudName)
	if err != nil {
		panic(fmt.Sprintf(
			"invalid cloud name '%s' specified, cannot continue\n", cloudName))
	}
	return &env
}

func getKeyvaultAuthorizer(tenantID, clientID, clientSecret string) (autorest.Authorizer, error) {
	// BUG: default value for KeyVaultEndpoint is wrong
	vaultEndpoint := strings.TrimSuffix(environment().KeyVaultEndpoint, "/")
	// BUG: alternateEndpoint replaces other endpoints in the configs below
	alternateEndpoint, _ := url.Parse("https://login.windows.net/" + tenantID + "/oauth2/token")

	var a autorest.Authorizer
	var err error

	oauthconfig, err := adal.NewOAuthConfig(environment().ActiveDirectoryEndpoint, tenantID)
	if err != nil {
		return a, err
	}
	oauthconfig.AuthorizeEndpoint = *alternateEndpoint

	token, err := adal.NewServicePrincipalToken(
		*oauthconfig, clientID, clientSecret, vaultEndpoint)

	if err != nil {
		return a, err
	}

	keyvaultAuthorizer := autorest.NewBearerAuthorizer(token)

	return keyvaultAuthorizer, err
}

func getKeysClient(authorizer autorest.Authorizer) *keyvault.BaseClient {
	keyClient := keyvault.New()
	keyClient.Authorizer = authorizer
	keyClient.AddToUserAgent(kvClientUserAgent)
	return &keyClient
}

func parseKeyVaultKeyInfo(keyVaultKeyIdentifier string) (*KeyVaultKeyInfo, error) {
	r, _ := regexp.Compile("https?://(.+)\\.vault\\.azure\\.net/keys/([^\\/.]+)/?([^\\/.]*)")

	str := r.FindStringSubmatch(keyVaultKeyIdentifier)
	if len(str) < 4 {
		return &KeyVaultKeyInfo{}, fmt.Errorf("Expected a key identifier from Key Vault. e.g.: https://keyvaultname.vault.azure.net/keys/myKey/99d67321dd9841af859129cd5551a871")
	}

	info := KeyVaultKeyInfo{}
	info.vaultURL = fmt.Sprintf("https://%s.vault.azure.net", str[1])
	info.keyName = str[2]
	info.keyVersion = str[3]

	return &info, nil
}

func (e *EncryptionClient) getKeyOperationsParameters(value *string) keyvault.KeyOperationsParameters {
	parameters := keyvault.KeyOperationsParameters{}
	parameters.Algorithm = keyvault.RSA15
	parameters.Value = value
	return parameters
}

func (e *EncryptionClient) Encrypt(ctx context.Context, data []byte) (*string, error) {
	if len(data) == 0 {
		v := ""
		return &v, nil
	}

	encoded := base64.RawStdEncoding.EncodeToString(data)

	parameters := e.getKeyOperationsParameters(&encoded)
	result, err := e.kvClient.Encrypt(ctx, e.kvInfo.vaultURL, e.kvInfo.keyName, e.kvInfo.keyVersion, parameters)
	if err != nil {
		return nil, err
	}

	return result.Result, nil
}

func (e *EncryptionClient) Decrypt(ctx context.Context, data *string) ([]byte, error) {
	if data == nil || len(*data) == 0 {
		return make([]byte, 0), nil
	}

	parameters := e.getKeyOperationsParameters(data)
	result, err := e.kvClient.Decrypt(ctx, e.kvInfo.vaultURL, e.kvInfo.keyName, e.kvInfo.keyVersion, parameters)
	if err != nil {
		return nil, err
	}

	decoded, err := base64.RawStdEncoding.DecodeString(*result.Result)
	if err != nil {
		return nil, err
	}

	return decoded, nil
}

func (e *EncryptionClient) EncryptBytes(ctx context.Context, data []byte) ([]byte, error) {
	if len(data) == 0 {
		return data, nil
	}

	encoded := base64.RawStdEncoding.EncodeToString(data)

	parameters := e.getKeyOperationsParameters(&encoded)
	result, err := e.kvClient.Encrypt(ctx, e.kvInfo.vaultURL, e.kvInfo.keyName, e.kvInfo.keyVersion, parameters)
	if err != nil {
		return nil, err
	}

	return []byte(*result.Result), nil
}

func (e *EncryptionClient) DecryptBytes(ctx context.Context, data []byte) ([]byte, error) {
	if len(data) == 0 {
		return data, nil
	}

	str := string(data)

	parameters := e.getKeyOperationsParameters(&str)
	result, err := e.kvClient.Decrypt(ctx, e.kvInfo.vaultURL, e.kvInfo.keyName, e.kvInfo.keyVersion, parameters)
	if err != nil {
		return nil, err
	}

	decoded, err := base64.RawStdEncoding.DecodeString(*result.Result)
	if err != nil {
		return nil, err
	}

	return decoded, nil
}
