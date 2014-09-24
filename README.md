# attspeech

[![wercker status](https://app.wercker.com/status/1c102c5109b0f8f4ecfe8f24e8eb8fcd/m "wercker status")](https://app.wercker.com/project/bykey/1c102c5109b0f8f4ecfe8f24e8eb8fcd)

Go client for the [AT&T Speech API](http://developer.att.com/apis/speech).

## Installation

	go get bitbucket.org/jsgoecke/attspeech

## Documentation

[http://godoc.org/github.com/jsgoecke/attspeech](http://godoc.org/github.com/jsgoecke/attspeech)

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

### Speech to Text Custom Result

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

	apiRequest := client.NewAPIRequest(attspeech.STTCResource)
	apiRequest.ContentType = "audio/x-wav"
	apiRequest.Data = data
	apiRequest.Filename = "test.wav"
	response, err := client.SpeechToTextCustom(apiRequest, srgsXML(), plsXML())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(response)
}

func plsXML() string {
	return `<?xml version="1.0" encoding="UTF-8"?> 
			<lexicon version="1.0" alphabet="sampa" xml:lang="en-US"> 
			   <lexeme> 
			       <grapheme>star</grapheme> 
			       <phoneme>tS { n</phoneme> 
			   </lexeme> 
			</lexicon>`
}

func srgsXML() string {
	return `<grammar root="top" xml:lang="en-US"> 
			  <rule id="CONTACT"> 
			      <one-of> 
			        <item>star</item> 
			        <item>key</item> 
			      </one-of> 
			  </rule> 
			  <rule id="top" scope="public"> 
			      <item> 
			          <one-of> 
			            <item>greeting</item> 
			            <item>the administration menu</item> 
			          </one-of> 
			      </item> 
			  <ruleref uri="#CONTACT"/> 
			  </rule> 
			</grammar>`
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

### Test Coverage

[http://gocover.io/github.com/jsgoecke/attspeech](http://gocover.io/github.com/jsgoecke/attspeech)

### Lint

[http://go-lint.appspot.com/github.com/jsgoecke/attspeech](http://go-lint.appspot.com/github.com/jsgoecke/attspeech)