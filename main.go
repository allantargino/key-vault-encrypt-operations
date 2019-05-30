package main

import (
	"context"
	"fmt"
	"io/ioutil"
)

func main() {
	fmt.Printf("Start!\n")

	ctx := context.Background()

	azureConfiguration, err := ParseEnvironment()
	if err != nil {
		panic(err)
	}

	client, err := NewEncryptionClientFromEnv(azureConfiguration)
	if err != nil {
		panic(err)
	}

	data, err := ioutil.ReadFile("./state.tfstate")
	if err != nil {
		panic(err)
	}

	// c := 245
	// n := cap(data) / c
	// for i := 0; i < n; i++ {
	// 	data = data[0:c]
	// 	fmt.Printf("i: %v, bytes: %v, cap: %v \n", i, len(data), cap(data))

	// 	operate(ctx, client, data)

	// 	data = data[c:]
	// }
	// data = data[0:cap(data)]
	// fmt.Printf("i: %v, bytes: %v, cap: %v \n", n, len(data), cap(data))

	// operate(ctx, client, data)

	fmt.Printf("init: bytes: %v\n", len(data))

	c := 245
	n := len(data) / c
	for i := 0; i < n; i++ {
		d := data[i*c : (i+1)*c]
		// fmt.Printf("i: %v, bytes: %v \n", i, len(d))
		operate(ctx, client, d)
	}
	d := data[n*c : len(data)]
	// fmt.Printf("f: %v, bytes: %v \n", n, len(d))

	operate(ctx, client, d)
}

type operate func(context.Context, *EncryptionClient, []byte) []byte

func operate(ctx context.Context, client *EncryptionClient, data []byte) {
	encryptedData, err := client.EncryptBytes(ctx, data)
	if err != nil {
		panic(err)
	}
	// fmt.Printf("enc: %s\n", string(encryptedData))

	decryptedData, err := client.DecryptBytes(ctx, encryptedData)
	if err != nil {
		panic(err)
	}
	fmt.Printf("data: %v\n", string(decryptedData))
}
