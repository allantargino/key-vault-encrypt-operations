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

	client, err := NewEncryptionClientFromEnv(azureConfiguration)
	if err != nil {
		panic(err)
	}

	msg := "this a (not so) random text!"
	fmt.Printf("msg: %s\n", msg)

	encryptedText, err := client.Encrypt(ctx, []byte(msg))
	if err != nil {
		panic(err)
	}
	fmt.Printf("enc: %s\n", *encryptedText)

	decryptedText, err := client.Decrypt(ctx, encryptedText)
	if err != nil {
		panic(err)
	}

	msg2 := string(decryptedText)

	fmt.Printf("eq: %t\n", msg == msg2)
}
