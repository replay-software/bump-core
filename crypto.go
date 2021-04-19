package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

func signFileWithKey(key string, appFilename string) string {
	data, err := base64.StdEncoding.DecodeString(key)

	if err != nil {
		println("An error occurred while signing")
		fmt.Println(err)
	}

	publicKey := data[64:96]
	privateKey := data[0:64]

	lastSig := sign(privateKey, publicKey, readInput(appFilename))
	lastSigEnc := base64.StdEncoding.EncodeToString([]byte(lastSig))
	return lastSigEnc
}

func readInput(appFilename string) (fileData []byte) {
	appFilenameSegments := strings.Split(appFilename, ".")
	lastSegmentWithExtension := appFilenameSegments[len(appFilenameSegments)-1]

	fileData, err := ioutil.ReadFile(findFileWithExtension(fmt.Sprintf(".%s", lastSegmentWithExtension)))

	if err != nil {
		println("Signing failed")
		fmt.Fprintf(os.Stderr, "Failed to read input file: %v.\n", err)
		os.Exit(1)
	}

	return
}
