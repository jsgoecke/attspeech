package main

import (
	"fmt"
	"github.com/jsgoecke/attspeech"
	"io/ioutil"
	"os"
)

func main() {
	client := attspeech.New(os.Getenv("ATT_APP_KEY"), os.Getenv("ATT_APP_SECRET"), "")
	client.SetAuthTokens()
	apiRequest := client.NewAPIRequest(attspeech.TTSResource)
	apiRequest.ContentType = "text/plain"
	apiRequest.Accept = "audio/x-wav"
	apiRequest.Text = "I want to be an airborne ranger, I want to live the life of danger."
	data, err := client.TextToSpeech(apiRequest)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	err = ioutil.WriteFile("/Users/jsgoecke/Desktop/tts_test.wav", data, 0644)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
