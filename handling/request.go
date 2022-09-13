package handling

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/denizakturk/dispatcher/constants"
)

const ContentTypeApplicationJson = "application/json"

func RequestHandle(req *http.Request) ([]byte, error) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf(constants.REQUEST_BODY_READ_ERROR, err)
	}

	req.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	contentType := req.Header.Get("Content-Type")
	if contentType != ContentTypeApplicationJson {
		return nil, fmt.Errorf(constants.CONTENT_TYPE_NOT_JSON)
	}
	if !json.Valid(body) {
		return nil, fmt.Errorf(constants.THIS_REQUEST_TYPE_INVALID_JSON)
	}

	return body, nil
}
