package attspeech

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"reflect"
	"runtime"
	"strings"
	"unicode"
)

const (
	// APIBase is the base URL for the ATT Speech API
	APIBase = "https://api.att.com"
	// STTResource is the speech to text resource
	STTResource = "/speech/v3/speechToText"
	// STTCResource is the speech to text custom resource
	STTCResource = "/speech/v3/speechToTextCustom"
	// TTSResource is the text to speech resource
	TTSResource = "/speech/v3/textToSpeech"
	// OauthResource is the oauth resource
	OauthResource = "/oauth/access_token"
	// UserAgent is the user agent use for the HTTP client
	UserAgent = "GoATTSpeechLib"
	// Version is the version of the ATT Speech API
	Version = "0.1"
)

/*
New creates a new AttSpeechClient

	client := attspeech.New("<id>", "<secret>", "")
	client.SetAuthTokens()
*/
func New(id string, secret string, apiBase string) *Client {
	client := &Client{
		STTResource:   STTResource,
		STTCResource:  STTCResource,
		TTSResource:   TTSResource,
		OauthResource: OauthResource,
		ID:            id,
		Secret:        secret,
		Scope:         [3]string{"SPEECH", "STTC", "TTS"},
		TTSFields:     [3]string{"Volume", "Tempo", "VoiceName"},
	}
	if apiBase == "" {
		client.APIBase = APIBase
	} else {
		client.APIBase = apiBase
	}
	return client
}

/*
SetAuthTokens sets the provided authorization tokens for the client

	client := attspeech.New("<id>", "<secret>", "")
	client.SetAuthTokens()
*/
func (client *Client) SetAuthTokens() error {
	data := "grant_type=client_credentials&"
	data += "client_id=" + client.ID + "&"
	data += "client_secret=" + client.Secret + "&"
	data += "scope="

	m := make(map[string]*Token)
	for _, scope := range client.Scope {
		req, _ := http.NewRequest("POST", client.APIBase+OauthResource+"?"+data+scope, nil)
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		body, _ := ioutil.ReadAll(res.Body)
		res.Body.Close()
		token := &Token{}
		err = json.Unmarshal(body, token)
		if err != nil {
			return err
		}
		m[scope] = token
	}
	client.Tokens = m
	return nil
}

/*
SpeechToText converts an audio file to text

	client := attspeech.New("<id>", "<secret>", "")
	client.SetAuthTokens()
	apiRequest := client.NewAPIRequest(STTResource)
	apiRequest.Data = data // where data is audio content as *bytes.Buffer
	apiRequest.ContentType = "audio/wav"
	result, apiError, err := client.SpeechToText(apiRequest)

More details available here:

	http://developer.att.com/apis/speech/docs#resources-speech-to-text
*/
func (client *Client) SpeechToText(apiRequest *APIRequest) (*Recognition, error) {
	if apiRequest.ContentType == "" {
		return nil, errors.New("A ContentType must be provided")
	}
	if apiRequest.Data == nil {
		return nil, errors.New("Data to convert to text must be provided")
	}

	body, statusCode, err := client.post(client.STTResource, apiRequest.Data, apiRequest)
	if err != nil {
		return nil, err
	}
	if statusCode == 200 {
		recognition := &Recognition{}
		json.Unmarshal(body, recognition)
		return recognition, nil
	}
	apiError := &APIError{}
	json.Unmarshal(body, apiError)
	return nil, apiError.generateErr()
}

/*
SpeechToTextCustom converts an audio file to text

	client := attspeech.New("<id>", "<secret>", "")
	client.SetAuthTokens()
	apiRequest := client.NewAPIRequest(STTResource)
	apiRequest.Data = data // where data is audio content as *bytes.Buffer
	apiRequest.ContentType = "audio/wav"
	apiRequest.Filename = "test.wav"
	result, apiError, err := client.SpeechToTextCustom(apiRequest, "<some srgs XML>", "<some pls XML>")

More details available here:

	http://developer.att.com/apis/speech/docs#resources-speech-to-text-custom

*/
func (client *Client) SpeechToTextCustom(apiRequest *APIRequest, grammar string, dictionary string) (*Recognition, error) {
	if grammar == "" {
		return nil, errors.New("A grammar must be provided")
	}
	if apiRequest.Data == nil {
		return nil, errors.New("Data must be provided")
	}
	if apiRequest.Filename == "" {
		return nil, errors.New("Filename must be provided")
	}
	if apiRequest.ContentType == "" {
		return nil, errors.New("ContentType must be provided")
	}

	apiRequest.Data, apiRequest.ContentType = buildForm(apiRequest, grammar, dictionary)
	body, statusCode, err := client.post(client.STTCResource, apiRequest.Data, apiRequest)
	if err != nil {
		return nil, err
	}
	if statusCode == 200 {
		recognition := &Recognition{}
		json.Unmarshal(body, recognition)
		return recognition, nil
	}
	apiError := &APIError{}
	return nil, apiError.generateErr()
}

/*
TextToSpeech converts text to a speech file

	client := attspeech.New("<id>", "<secret>", "")
	client.SetAuthTokens()

	apiRequest := client.NewAPIRequest(TTSResource)
	apiRequest.Accept = "audio/x-wav",
	apiRequest.VoiceName = "crystal",
	apiRequest.Text = "I want to be an airborne ranger, I want to live the life of danger.",
	data, err := client.TextToSpeech(apiRequest)

More details available here:

	http://developer.att.com/apis/speech/docs#resources-text-to-speech
*/
func (client *Client) TextToSpeech(apiRequest *APIRequest) ([]byte, error) {
	if apiRequest.Text == "" {
		return nil, errors.New("Text to convert to speech must be provided")
	}

	body, statusCode, err := client.post(client.TTSResource, bytes.NewBuffer([]byte(apiRequest.Text)), apiRequest)
	if err != nil {
		return nil, err
	}
	if statusCode == 200 {
		return body, nil
	}
	apiError := &APIError{}
	json.Unmarshal(body, apiError)
	return nil, apiError.generateErr()
}

/*
NewAPIRequest sets the common headers for TTS and STT

	client := attspeech.New("<id>", "<secret>", "")
	client.SetAuthTokens()
	apiRequest := client.NewAPIRequest(TTSResource)

Note, when setting apiRequest.XArg, always append with '+=', unless you specifically
intend to overwrite the defaults for ClientApp, ClientVersion, DeviceType and DeviceOs
*/
func (client *Client) NewAPIRequest(resource string) *APIRequest {
	apiRequest := &APIRequest{}
	apiRequest.Accept = "application/json"
	apiRequest.UserAgent = "Golang net/http"
	apiRequest.XArg = "ClientApp=GoLibForATTSpeech,"
	apiRequest.XArg += "ClientVersion=" + Version + ","
	apiRequest.XArg += "DeviceType=" + runtime.GOARCH + ","
	apiRequest.XArg += "DeviceOs=" + runtime.GOOS

	switch resource {
	case client.STTResource:
		apiRequest.Authorization = "Bearer " + client.Tokens["SPEECH"].AccessToken
		apiRequest.TransferEncoding = "chunked"
	case client.STTCResource:
		apiRequest.Authorization = "Bearer " + client.Tokens["STTC"].AccessToken
	case client.TTSResource:
		apiRequest.Authorization = "Bearer " + client.Tokens["TTS"].AccessToken
		apiRequest.ContentType = "text/plain"
	case client.OauthResource:
		apiRequest.ContentType = "application/x-www-form-urlencoded"
	}
	return apiRequest
}

// post to the AT&T Speech API
func (client *Client) post(resource string, body *bytes.Buffer, apiRequest *APIRequest) ([]byte, int, error) {
	req, err := http.NewRequest("POST", client.APIBase+resource, body)
	if err != nil {
		return nil, 0, err
	}
	apiRequest.setHeaders(req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	respBody, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	return respBody, resp.StatusCode, nil
}

// generateErr takes the APIError and turns it into a Go error
func (apiError *APIError) generateErr() error {
	msg := apiError.RequestError.ServiceException.MessageID + " - "
	msg += apiError.RequestError.ServiceException.Text + " - "
	msg += apiError.RequestError.ServiceException.Variables
	if msg == " -  - " {
		msg = apiError.RequestError.PolicyException.MessageID + " - "
		msg += apiError.RequestError.PolicyException.Text + " - "
		msg += apiError.RequestError.PolicyException.Variables

	}
	if msg == " -  - " {
		/*
			#FIXME
			http://developerboards.att.lithium.com/t5/API-Platform/Speech-API-STTC-Error-Returns-Invalid-JSON/td-p/38929
		*/
		msg = "Could not parse JSON error from the AT&T Speech API"
	}
	return errors.New(msg)
}

// setHeaders returns the APIRequest as a map
func (apiRequest *APIRequest) setHeaders(req *http.Request) {
	headers := make(map[string]string)
	xarg := ""

	s := reflect.ValueOf(apiRequest).Elem()
	typeOfT := s.Type()
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		name := typeOfT.Field(i).Name
		if name != "Data" && name != "Text" && name != "Filename" {
			if name == "VoiceName" || name == "Volume" || name == "Tempo" {
				if f.Interface().(string) != "" {
					xarg += "," + name + "=" + f.Interface().(string)
				}
			} else {
				headers[toDash(name)] = f.Interface().(string)
			}
		}
	}
	headers["X-Arg"] += xarg
	for key, value := range headers {
		req.Header.Add(key, value)
	}
}

/*
toDash converts an uppercase string into a string
where uppercase letters are sperated by a '-'
*/
func toDash(value string) string {
	var words []string
	l := 0
	for s := value; s != ""; s = s[l:] {
		l = strings.IndexFunc(s[1:], unicode.IsUpper) + 1
		if l <= 0 {
			l = len(s)
		}
		words = append(words, s[:l])
	}
	dashedWord := ""
	numWords := len(words)
	for i := 0; i < numWords; i++ {
		if i == 0 && numWords > 1 {
			dashedWord = words[0] + "-"
		} else {
			dashedWord += words[i]
		}
	}
	return dashedWord
}

//buildForm builds a multipart form to send the file with
func buildForm(apiRequest *APIRequest, grammar string, dictionary string) (*bytes.Buffer, string) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	if dictionary != "" {
		// Add the dictionary field
		contentDisposition := "form-data; name=\"x-dictionary\"; filename=\"speech_alpha.pls\""
		apiRequest.addField(writer, dictionary, contentDisposition, "application/pls+xml")
	}

	// Add the grammar field
	contentDisposition := "form-data; name=\"x-grammar\""
	apiRequest.addField(writer, grammar, contentDisposition, "application/srgs+xml")

	// Add the file field
	contentDisposition = "form-data; name=\"x-voice\"; filename=\"" + apiRequest.Filename + "\""
	apiRequest.addField(writer, grammar, contentDisposition, apiRequest.ContentType)

	writer.Close()

	contentType := writer.FormDataContentType()
	contentType = strings.Replace(contentType, "form-data", "x-srgs-audio", 1)
	return body, contentType
}

func (apiRequest *APIRequest) addField(writer *multipart.Writer, body string, contentDisposition string, contentType string) {
	header := make(map[string][]string)
	header["Content-Disposition"] = []string{contentDisposition}
	header["Content-Type"] = []string{contentType}
	part, _ := writer.CreatePart(header)
	if contentType == "application/pls+xml" || contentType == "application/srgs+xml" {
		part.Write([]byte(body + "\n"))
	} else {
		io.Copy(part, apiRequest.Data)
	}
}
