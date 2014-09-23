package attspeech

import (
	"encoding/json"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

const (
	RecognitionJSON = `
	{
	    "Recognition": {
	        "Status": "Ok",
	        "ResponseId": "3125ae74122628f44d265c231f8fc926",
	        "NBest": [
	            {
	                "Hypothesis": "bookstores in glendale california",
	                "LanguageId": "en-us",
	                "Confidence": 0.9,
	                "Grade": "accept",
	                "ResultText": "bookstores in Glendale, CA",
	                "Words": [
	                    "bookstores",
	                    "in",
	                    "glendale",
	                    "california"
	                ],
	                "WordScores": [
	                    0.92,
	                    0.73,
	                    0.81,
	                    0.96
	                ]
	            }
	        ]
	    }
	}`

	TokenJSON = `
	{
	    "access_token":"1234",
	    "token_type": "bearer",
	    "expires_in":5678,
	    "refresh_token":"9abc"
	}`
)

func TestRecognition(t *testing.T) {
	Convey("A valid recognition object should be returned", t, func() {
		recognition := &Recognition{}
		err := json.Unmarshal([]byte(RecognitionJSON), recognition)

		So(err, ShouldBeNil)
		So(recognition.Recognition.Status, ShouldEqual, "Ok")
		So(recognition.Recognition.ResponseID, ShouldEqual, "3125ae74122628f44d265c231f8fc926")
		So(recognition.Recognition.NBest[0].LanguageID, ShouldEqual, "en-us")
	})
}

func TestToken(t *testing.T) {
	Convey("A valid token object should be returned", t, func() {
		token := &Token{}
		err := json.Unmarshal([]byte(TokenJSON), token)

		So(err, ShouldBeNil)
		So(token.AccessToken, ShouldEqual, "1234")
		So(token.TokenType, ShouldEqual, "bearer")
		So(token.ExpiresIn, ShouldEqual, 5678)
		So(token.RefreshToken, ShouldEqual, "9abc")
	})
}
