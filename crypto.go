package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
)

func signFileWithKey(key string) string {
	data, err := base64.StdEncoding.DecodeString(key)

	if err != nil {
		println("An error occurred while signing")
		fmt.Println(err)
	}

	publicKey := data[64:96]
	privateKey := data[0:64]

	lastSig := sign(privateKey, publicKey, readInput())
	lastSigEnc := base64.StdEncoding.EncodeToString([]byte(lastSig))
	return lastSigEnc
}

func readInput() (fileData []byte) {
	fileData, err := ioutil.ReadFile(findFileWithExtension(".zip"))

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read input file: %v.\n", err)
		os.Exit(1)
	}

	return
}
