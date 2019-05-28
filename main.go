package main

import (
	"context"
	"fmt"
	"net/url"
	"strings"

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

func getSecret(ctx context.Context, vaultURL, keyName string, authorizer autorest.Authorizer) (key keyvault.SecretBundle, err error) {
	keyClient := getKeysClient(authorizer)
	return keyClient.GetSecret(ctx, vaultURL, keyName, "")
}
