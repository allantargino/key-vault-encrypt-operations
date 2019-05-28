package main

import (
	"context"
	"fmt"
)

func main() {
	println("Start!")

	ctx := context.Background()

	azureConfiguration, err := ParseEnvironment()
	if err != nil {
		panic(err)
	}

	client, err := NewEncryptionClient(azureConfiguration)
	if err != nil {
		panic(err)
	}

	msg := "this is something to encrypt!"
	fmt.Println(msg)

	encryptedText, err := client.Encrypt(ctx, []byte(msg))
	if err != nil {
		panic(err)
	}
	fmt.Println(*encryptedText)

	decryptedText, err := client.Decrypt(ctx, encryptedText)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(decryptedText))
}
