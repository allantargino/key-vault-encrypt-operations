package main

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"encoding/base64"

	"github.com/Azure/azure-sdk-for-go/services/keyvault/2016-10-01/keyvault"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
)

func main() {
	println("Start!")

	ctx := context.Background()

	azureConfiguration, err := ParseEnvironment()
	if err != nil {
		panic(err)
	}

	authorizer, err := getKeyvaultAuthorizer(azureConfiguration)

	if err != nil {
		panic(err)
	}

	secret, err := getSecret(ctx, azureConfiguration.KeyVaultUrl, "secretName", authorizer)
	if err != nil {
		panic(err)
	}
	fmt.Println(*secret.Value)

	msg := "aijoieajdoiwjdoiwaj!"
	fmt.Println(msg)
	encoded := base64.RawStdEncoding.EncodeToString([]byte(msg))
	fmt.Println(encoded)

	encryptedText, err := encrypt(ctx, azureConfiguration.KeyVaultUrl, "myKey", authorizer, &encoded)
	if err != nil {
		panic(err)
	}
	fmt.Println(*encryptedText.Result)

	decryptedText, err := decrypt(ctx, azureConfiguration.KeyVaultUrl, "myKey", authorizer, encryptedText.Result)
	if err != nil {
		panic(err)
	}
	fmt.Println(*decryptedText.Result)
	
	decoded, err := base64.RawStdEncoding.DecodeString(*decryptedText.Result)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(decoded))
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
	keyClient.AddToUserAgent("azure-samples")
	return keyClient
}

func getSecret(ctx context.Context, vaultURL, secretName string, authorizer autorest.Authorizer) (key keyvault.SecretBundle, err error) {
	keyClient := getKeysClient(authorizer)
	return keyClient.GetSecret(ctx, vaultURL, secretName, "")
}

func encrypt(ctx context.Context, vaultURL, keyName string, authorizer autorest.Authorizer, value *string) (key keyvault.KeyOperationResult, err error) {
	keyClient := getKeysClient(authorizer)
	parameters := keyvault.KeyOperationsParameters{}
	parameters.Algorithm = keyvault.RSAOAEP256
	parameters.Value = value

	return keyClient.Encrypt(ctx, vaultURL, keyName, "", parameters)
}

func decrypt(ctx context.Context, vaultURL, keyName string, authorizer autorest.Authorizer, value *string) (key keyvault.KeyOperationResult, err error) {
	keyClient := getKeysClient(authorizer)
	parameters := keyvault.KeyOperationsParameters{}
	parameters.Algorithm = keyvault.RSAOAEP256
	parameters.Value = value

	return keyClient.Decrypt(ctx, vaultURL, keyName, "", parameters)
}
