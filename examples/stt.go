package main

import (
	"bytes"
	"fmt"
	"github.com/jsgoecke/attspeech"
	"io"
	"os"
)

func main() {
	client := attspeech.New(os.Getenv("ATT_APP_KEY"), os.Getenv("ATT_APP_SECRET"), "")
	client.SetAuthTokens()

	data := &bytes.Buffer{}
	file, err := os.Open("../test/test.wav")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer file.Close()
	_, err = io.Copy(data, file)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	apiRequest := client.NewAPIRequest(attspeech.STTResource)
	apiRequest.ContentType = "audio/x-wav"
	apiRequest.Data = data
	response, err := client.SpeechToText(apiRequest)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(response)
}
