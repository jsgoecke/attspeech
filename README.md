# attspeech

Go client for the [AT&T Speech API](http://developer.att.com/apis/speech).

## Installation

	go get bitbucket.org/jsgoecke/attspeech

## Usage

### Speech to Text Result

```go
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
```

### Text to Speech Result

```go
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
```

## Testing
	
	cd attspeech
	go get github.com/smartystreets/goconvey/convey
	go test

## ToDo

	* Implement Custom Speech to Text
