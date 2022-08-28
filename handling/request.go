package handling

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

const ContentTypeApplicationJson = "application/json"

func RequestHandle(req *http.Request) ([]byte, error) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Printf("Error Reading Body %v", err)
		return nil, err
	}

	req.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	contentType := req.Header.Get("Content-Type")

	if contentType != ContentTypeApplicationJson || !json.Valid(body) {
		//var rw http.ResponseWriter
		//http.Error(rw, "Invalid Document", 500)
		return nil, err
	}

	return body, nil
}
