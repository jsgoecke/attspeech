package attspeech

import (
	"bytes"
	"encoding/json"
	. "github.com/smartystreets/goconvey/convey"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestClient(t *testing.T) {
	Convey("Creating a new client should set values correctly", t, func() {
		client := New("foo", "bar", "")
		So(client.APIBase, ShouldEqual, "https://api.att.com")
		So(client.ID, ShouldEqual, "foo")
		So(client.Secret, ShouldEqual, "bar")
		So(client.STTResource, ShouldEqual, STTResource)
		So(client.STTCResource, ShouldEqual, STTCResource)
		So(client.TTSResource, ShouldEqual, TTSResource)
		So(client.Scope, ShouldEqual, [3]string{"SPEECH", "STTC", "TTS"})
	})
}

func TestCustomApiBase(t *testing.T) {
	Convey("Creating a new client with a custom API URL should set values correctly", t, func() {
		client := New("foo", "bar", "http://foobar.com")
		So(client.APIBase, ShouldEqual, "http://foobar.com")
	})
}

func TestNewAPIRequest(t *testing.T) {
	ts := serveHTTP(t)
	client := New("foo", "bar", "")
	client.APIBase = ts.URL
	client.SetAuthTokens()
	Convey("Should generate a new APIRequest object", t, func() {
		Convey("STTResource", func() {
			apiRequest := client.NewAPIRequest(client.STTResource)
			So(apiRequest.TransferEncoding, ShouldEqual, "chunked")
			So(apiRequest.Authorization, ShouldEqual, "Bearer "+client.Tokens["SPEECH"].AccessToken)
		})
		Convey("STTCResource", func() {
			apiRequest := client.NewAPIRequest(client.STTCResource)
			So(apiRequest.Authorization, ShouldEqual, "Bearer "+client.Tokens["STTC"].AccessToken)
		})
		Convey("TTSResource", func() {
			apiRequest := client.NewAPIRequest(client.TTSResource)
			So(apiRequest.ContentType, ShouldEqual, "text/plain")
			So(apiRequest.Authorization, ShouldEqual, "Bearer "+client.Tokens["TTS"].AccessToken)
		})
		Convey("OauthResource", func() {
			apiRequest := client.NewAPIRequest(client.OauthResource)
			So(apiRequest.ContentType, ShouldEqual, "application/x-www-form-urlencoded")
		})
	})
}

func TestToDash(t *testing.T) {
	Convey("Converting struct elements to HTTP headers", t, func() {
		Convey("Should leave this word undashed", func() {
			word := toDash("Foobar")
			So(word, ShouldEqual, "Foobar")
		})
		Convey("Should put one dash in this word", func() {
			word := toDash("FooBar")
			So(word, ShouldEqual, "Foo-Bar")
		})
		Convey("Should put one dash in this word with multiple caps", func() {
			word := toDash("FooBarBaz")
			So(word, ShouldEqual, "Foo-BarBaz")
		})
	})
}

func TestSetHeaders(t *testing.T) {
	Convey("Should handle X-Arg additions properly", t, func() {
		ts := serveHTTP(t)
		client := New("foo", "bar", "")
		client.APIBase = ts.URL
		client.SetAuthTokens()
		apiRequest := client.NewAPIRequest(TTSResource)
		apiRequest.ContentType = "audio/x-wav"
		apiRequest.Text = "foobar"
		apiRequest.VoiceName = "alberto"
		apiRequest.Volume = "100"
		apiRequest.Tempo = "0"
		req, _ := http.NewRequest("POST", client.APIBase, nil)
		Convey("Should add VoiceName, Volume and Temp to X-Arg", func() {
			apiRequest.setHeaders(req)
			So(req.Header.Get("X-Arg"), ShouldEqual, "ClientApp=GoLibForATTSpeech,ClientVersion=0.1,DeviceType=amd64,DeviceOs=darwin,Tempo=0,VoiceName=alberto,Volume=100")
		})
		Convey("Should add additional X-Arg params while preserving the original ones", func() {
			apiRequest.XArg += ",ShowWordTokens=true"
			apiRequest.setHeaders(req)
			So(req.Header.Get("X-Arg"), ShouldEqual, "ClientApp=GoLibForATTSpeech,ClientVersion=0.1,DeviceType=amd64,DeviceOs=darwin,ShowWordTokens=true,Tempo=0,VoiceName=alberto,Volume=100")
		})
		Convey("Should only render headers that have values", func() {
			apiRequest.XArg += ",ShowWordTokens=true"
			apiRequest.ContentType = ""
			apiRequest.setHeaders(req)
			So(req.Header.Get("ContentType"), ShouldBeBlank)
		})
	})
}

func TestGetTokens(t *testing.T) {
	Convey("Should get proper tokens", t, func() {
		ts := serveHTTP(t)
		client := New("foo", "bar", "")
		client.APIBase = ts.URL
		err := client.SetAuthTokens()

		So(err, ShouldBeNil)
		scopes := [3]string{"SPEECH", "TTS", "STTC"}
		for _, scope := range scopes {
			So(client.Tokens[scope].AccessToken, ShouldEqual, "123")
			So(client.Tokens[scope].RefreshToken, ShouldEqual, "456")
		}
	})
}

func TestSpeechToText(t *testing.T) {
	Convey("Should return a recognition of an audio file", t, func() {
		Convey("When no ContentType is provided", func() {
			ts := serveHTTP(t)
			client := New(os.Getenv("ATT_APP_KEY"), os.Getenv("ATT_APP_SECRET"), "")
			client.APIBase = ts.URL
			client.SetAuthTokens()
			apiRequest := client.NewAPIRequest(STTResource)
			response, err := client.SpeechToText(apiRequest)
			So(response, ShouldBeNil)
			So(err.Error(), ShouldEqual, "a content type must be provided")
		})
		Convey("When no Data is provided", func() {
			ts := serveHTTP(t)
			client := New(os.Getenv("ATT_APP_KEY"), os.Getenv("ATT_APP_SECRET"), "")
			client.APIBase = ts.URL
			client.SetAuthTokens()
			apiRequest := client.NewAPIRequest(STTResource)
			apiRequest.ContentType = "audio/x-wav"
			response, err := client.SpeechToText(apiRequest)
			So(response, ShouldBeNil)
			So(err.Error(), ShouldEqual, "data to convert to text must be provided")
		})
		Convey("When an invalid ContentType is provided", func() {
			ts := serveHTTP(t)
			client := New(os.Getenv("ATT_APP_KEY"), os.Getenv("ATT_APP_SECRET"), "")
			client.APIBase = ts.URL
			client.SetAuthTokens()

			// Read the test file
			data := &bytes.Buffer{}
			file, err := os.Open("./test/test.wav")
			So(err, ShouldBeNil)
			defer file.Close()
			_, err = io.Copy(data, file)
			So(err, ShouldBeNil)

			apiRequest := client.NewAPIRequest(STTResource)
			apiRequest.Data = data
			apiRequest.ContentType = "foo/bar"

			response, err := client.SpeechToText(apiRequest)

			So(response, ShouldBeNil)
			So(err.Error(), ShouldEqual, "SVC0002 - Invalid input value for message part %1 - Content-Type")
		})
		Convey("When a valid ContentType is provided", func() {
			ts := serveHTTP(t)
			client := New(os.Getenv("ATT_APP_KEY"), os.Getenv("ATT_APP_SECRET"), "")
			client.APIBase = ts.URL
			client.SetAuthTokens()

			// Read the test file
			data := &bytes.Buffer{}
			file, err := os.Open("./test/test.wav")
			So(err, ShouldBeNil)
			defer file.Close()
			_, err = io.Copy(data, file)
			So(err, ShouldBeNil)

			apiRequest := client.NewAPIRequest(STTResource)
			apiRequest.Data = data
			apiRequest.ContentType = "audio/wav"
			response, err := client.SpeechToText(apiRequest)

			So(err, ShouldBeNil)
			So(response.Recognition.Status, ShouldEqual, "OK")
			So(response.Recognition.NBest[0].ResultText, ShouldEqual, "If you wish to keep this new greeting press one if you wish to record the greeting press two. To re store your old greeting in return to the administration menu. Press the star key.")
		})
	})
}

func TestBuildForm(t *testing.T) {
	ts := serveHTTP(t)
	client := New(os.Getenv("ATT_APP_KEY"), os.Getenv("ATT_APP_SECRET"), "")
	client.APIBase = ts.URL
	client.SetAuthTokens()
	apiRequest := client.NewAPIRequest(STTCResource)
	apiRequest.ContentType = "audio/x-wav"
	apiRequest.Filename = "test.wav"
	apiRequest.Data = bytes.NewBuffer([]byte(`foobar`))
	Convey("Should build a multipart form", t, func() {
		Convey("With a dictionary field", func() {
			body, contentType := buildForm(apiRequest, "<foo>bar</foo>", "<baz>bar</baz>")
			So(strings.Contains(contentType, "multipart/x-srgs-audio"), ShouldBeTrue)
			bodyStr := body.String()
			So(strings.Contains(bodyStr, "application/pls+xml"), ShouldBeTrue)
			So(strings.Contains(bodyStr, "application/srgs+xml"), ShouldBeTrue)
			So(strings.Contains(bodyStr, "audio/x-wav"), ShouldBeTrue)
		})
		Convey("Without a dictionary field", func() {
			body, contentType := buildForm(apiRequest, "<foo>bar</foo>", "")
			So(strings.Contains(contentType, "multipart/x-srgs-audio"), ShouldBeTrue)
			bodyStr := body.String()
			So(strings.Contains(bodyStr, "application/pls+xml"), ShouldBeFalse)
			So(strings.Contains(bodyStr, "application/srgs+xml"), ShouldBeTrue)
			So(strings.Contains(bodyStr, "audio/x-wav"), ShouldBeTrue)
		})
	})
}

func TestSpeechToTextCustom(t *testing.T) {
	Convey("Should return a recognition of an audio file", t, func() {
		ts := serveHTTP(t)
		client := New(os.Getenv("ATT_APP_KEY"), os.Getenv("ATT_APP_SECRET"), "")
		client.APIBase = ts.URL
		client.SetAuthTokens()
		apiRequest := client.NewAPIRequest(STTCResource)

		Convey("When no Grammar is provided", func() {
			response, err := client.SpeechToTextCustom(apiRequest, "", "")
			So(response, ShouldBeNil)
			So(err.Error(), ShouldEqual, "a grammar must be provided")
		})
		Convey("When no Data is provided", func() {
			apiRequest.Data = nil
			response, err := client.SpeechToTextCustom(apiRequest, "foobar", "")
			So(response, ShouldBeNil)
			So(err.Error(), ShouldEqual, "data must be provided")
		})

		apiRequest.Data = bytes.NewBuffer([]byte(`foobar`))

		Convey("When no Filename is provided", func() {
			response, err := client.SpeechToTextCustom(apiRequest, "foobar", "")
			So(response, ShouldBeNil)
			So(err.Error(), ShouldEqual, "filename must be provided")
		})
		Convey("When no ContentType is provided", func() {
			apiRequest.Filename = "foobar.wav"
			response, err := client.SpeechToTextCustom(apiRequest, "foobar", "")
			So(response, ShouldBeNil)
			So(err.Error(), ShouldEqual, "content type must be provided")
		})
		Convey("Should process a custom STT request", func() {
			// Read the test file
			data := &bytes.Buffer{}
			file, err := os.Open("./test/test.wav")
			So(err, ShouldBeNil)
			defer file.Close()
			_, err = io.Copy(data, file)
			So(err, ShouldBeNil)

			apiRequest.Data = data
			apiRequest.Filename = "test.wav"
			apiRequest.ContentType = "audio/wav"
			response, err := client.SpeechToTextCustom(apiRequest, srgsXML(), plsXML())
			So(err, ShouldBeNil)
			So(response.Recognition.Status, ShouldEqual, "OK")
			So(response.Recognition.ResponseID, ShouldEqual, "c7a420e9cdc50645412311b7c0365e34")
		})
	})
}

func TestTextToSpeech(t *testing.T) {
	Convey("Should handle Text to Speech (TTS)", t, func() {
		ts := serveHTTP(t)
		client := New(os.Getenv("ATT_APP_KEY"), os.Getenv("ATT_APP_SECRET"), "")
		client.APIBase = ts.URL
		client.SetAuthTokens()
		Convey("Should set the default ContentType", func() {
			apiRequest := client.NewAPIRequest(TTSResource)
			apiRequest.Text = "foobar"
			So(apiRequest.ContentType, ShouldEqual, "text/plain")
		})
		Convey("Should return an error if Text not set", func() {
			apiRequest := client.NewAPIRequest(TTSResource)
			_, err := client.TextToSpeech(apiRequest)
			So(err.Error(), ShouldEqual, "text to convert to speech must be provided")
		})
		Convey("Should return an error if an invalid ContentType", func() {
			apiRequest := client.NewAPIRequest(TTSResource)
			apiRequest.ContentType = "foo/bar"
			apiRequest.Text = "foobar"
			response, err := client.TextToSpeech(apiRequest)
			So(response, ShouldBeNil)
			So(err.Error(), ShouldEqual, "SVC0002 - Invalid input value for message part %1 - Content-Type")
		})
	})
}

func TestGenerateErr(t *testing.T) {
	Convey("Should generate error messages", t, func() {
		Convey("ServiceException", func() {
			apiError := &APIError{}
			json.Unmarshal(contentTypeErrorJSON(), apiError)
			err := apiError.generateErr()
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "SVC0002 - Invalid input value for message part %1 - Content-Type")
		})
		Convey("PolicyException", func() {
			apiError := &APIError{}
			json.Unmarshal(policyErrorJSON(), apiError)
			err := apiError.generateErr()
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "SVC0002 - Policy error - Content-Type")
		})
	})
}

func serveHTTP(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if strings.Contains(req.RequestURI, OauthResource) {
			w.WriteHeader(200)
			w.Write(oauthJSON())
			return
		}
		if strings.Contains(req.RequestURI, STTCResource) {
			checkHeaders(t, req)
			w.WriteHeader(200)
			w.Write(customRecoginitionJSON())
			return
		}
		if strings.Contains(req.RequestURI, STTResource) {
			checkHeaders(t, req)
			if req.Header.Get("Content-Type") == "foo/bar" {
				w.WriteHeader(400)
				w.Write(contentTypeErrorJSON())
				return
			}
			w.WriteHeader(200)
			w.Write(recognitionJSON())
			return
		}
		if strings.Contains(req.RequestURI, TTSResource) {
			checkHeaders(t, req)
			if req.Header.Get("Content-Type") == "foo/bar" {
				w.WriteHeader(400)
				w.Write(contentTypeErrorJSON())
				return
			}
			data, _ := ioutil.ReadFile("./test/tts_test.wav")
			w.WriteHeader(200)
			w.Write(data)
			return
		}
	}))
}

func checkHeaders(t *testing.T, req *http.Request) {
	Convey("Default headers should be set", t, func() {
		So(req.Header.Get("X-Arg"), ShouldEqual, "ClientApp=GoLibForATTSpeech,ClientVersion=0.1,DeviceType=amd64,DeviceOs=darwin")
		So(req.Header.Get("User-Agent"), ShouldEqual, "Golang net/http")
		So(req.Header.Get("Accept"), ShouldNotBeNil)
		So(req.Header.Get("Authorization"), ShouldEqual, "Bearer 123")
	})
}

func oauthJSON() []byte {
	return []byte(`
		{
		    "access_token":"123",
		    "token_type": "bearer",
		    "expires_in":500,
		    "refresh_token":"456"
		}
	`)
}

func recognitionJSON() []byte {
	return []byte(`
		{
		    "Recognition": {
		        "Info": {
		            "metrics": {
		                "audioBytes": 92102,
		                "audioTime": 11.5100002
		            }
		        },
		        "NBest": [
		            {
		                "Confidence": 0.667999999,
		                "Grade": "accept",
		                "Hypothesis": "if you wish to keep this new greeting press one if you wish to record the greeting press two to re store your old greeting in return to the administration menu press the star key",
		                "LanguageId": "en-US",
		                "ResultText": "If you wish to keep this new greeting press one if you wish to record the greeting press two. To re store your old greeting in return to the administration menu. Press the star key.",
		                "WordScores": [
		                    1,
		                    1,
		                    1,
		                    1,
		                    1,
		                    1,
		                    0.449,
		                    0.449,
		                    0.449,
		                    1,
		                    1,
		                    1,
		                    1,
		                    1,
		                    0.289,
		                    0.23,
		                    0.31,
		                    0.37,
		                    1,
		                    1,
		                    0.36,
		                    0.15,
		                    0.14,
		                    0.189,
		                    0.589,
		                    0.07,
		                    0.189,
		                    0.37,
		                    0.4,
		                    0.37,
		                    1,
		                    1,
		                    1,
		                    1,
		                    1
		                ],
		                "Words": [
		                    "If",
		                    "you",
		                    "wish",
		                    "to",
		                    "keep",
		                    "this",
		                    "new",
		                    "greeting",
		                    "press",
		                    "one",
		                    "if",
		                    "you",
		                    "wish",
		                    "to",
		                    "record",
		                    "the",
		                    "greeting",
		                    "press",
		                    "two.",
		                    "To",
		                    "re",
		                    "store",
		                    "your",
		                    "old",
		                    "greeting",
		                    "in",
		                    "return",
		                    "to",
		                    "the",
		                    "administration",
		                    "menu.",
		                    "Press",
		                    "the",
		                    "star",
		                    "key."
		                ]
		            }
		        ],
		        "ResponseId": "cf928a1adb259abf409da1993543fcdc",
		        "Status": "OK"
		    }
		}
	`)
}

func contentTypeErrorJSON() []byte {
	return []byte(`
	{
	    "RequestError": {
	        "ServiceException": {
	            "MessageId": "SVC0002",
	            "Text": "Invalid input value for message part %1",
	            "Variables": "Content-Type"
	        }
	    }
	}
	`)
}

func policyErrorJSON() []byte {
	return []byte(`
	{
	    "RequestError": {
	        "PolicyException": {
	            "MessageId": "SVC0002",
	            "Text": "Policy error",
	            "Variables": "Content-Type"
	        }
	    }
	}
	`)
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

func customRecoginitionJSON() []byte {
	return []byte(`
		{
		    "Recognition": {
		        "Info": {
		            "metrics": {
		                "audioBytes": 92187,
		                "audioTime": 11.5200005
		            }
		        },
		        "NBest": [
		            {
		                "Confidence": 0.78,
		                "Grade": "accept",
		                "Hypothesis": "greeting key",
		                "LanguageId": "en-US",
		                "ResultText": "greeting key",
		                "WordScores": [
		                    0.689,
		                    0.819
		                ],
		                "Words": [
		                    "greeting",
		                    "key"
		                ]
		            }
		        ],
		        "ResponseId": "c7a420e9cdc50645412311b7c0365e34",
		        "Status": "OK"
		    }
		}
	`)
}
