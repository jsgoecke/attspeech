package attspeech

import (
	"bytes"
	"regexp"
)

// Client is an ATT Speech API client
type Client struct {
	APIBase     string
	STTResource string
	TTSResource string
	ID          string
	Secret      string
	Tokens      map[string]*Token
	Scope       [3]string
	TTSFields   [3]string
}

// APIError represents an error from the AT&T Speech API
type APIError struct {
	RequestError struct {
		ServiceException struct {
			MessageID string `json:"MessageId"`
			Text      string `json:"Text"`
			Variables string `json:"Variables"`
		} `json:"ServiceException"`
		PolicyException struct {
			MessageID string `json:"MessageId"`
			Text      string `json:"Text"`
			Variables string `json:"Variables"`
		} `json:"PolicyException"`
	} `json:"RequestError"`
}

// ValueRegexes represents the regexes for all possible values of the API
type ValueRegexes struct {
	ContentType *regexp.Regexp
	FileType    *regexp.Regexp
	VoiceName   *regexp.Regexp
}

// Recognition represents at AT&T recognition response
type Recognition struct {
	Recognition struct {
		Status     string `json:"Status"`
		ResponseID string `json:"ResponseId"`
		NBest      []struct {
			Hypothesis    string    `json:"Hypothesis"`
			LanguageID    string    `json:"LanguageId"`
			Confidence    float32   `json:"Confidence"`
			Grade         string    `json:"Grade"`
			ResultText    string    `json:"ResultText"`
			Words         []string  `json:"Words"`
			WordScores    []float32 `json:"WordScores"`
			NluHypothesis struct {
				OutComposite []struct {
					Grammar string `json:"Grammar"`
					Out     string `json:"Out"`
				} `json:"OutComposite"`
			} `json:"NluHypothesis"`
		} `json:"NBest"`
	} `json:"Recognition"`
}

// Token represents the authorization tokens returned by the AT&T Speech API
type Token struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

// APIRequest represents the parameters for a Text to Speech request
type APIRequest struct {
	Authorization    string
	ContentType      string
	Accept           string
	VoiceName        string
	Text             string
	Volume           string
	Tempo            string
	TransferEncoding string
	UserAgent        string
	XArg             string
	Filename         string
	Data             *bytes.Buffer
}
