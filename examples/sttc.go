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
