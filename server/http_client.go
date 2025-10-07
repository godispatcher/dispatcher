package server

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/godispatcher/dispatcher/model"
)

// CallHTTP sends the given model.Document to a remote ServJsonApi endpoint
// hosted at http://host:port/ and returns the response document.
// It uses application/json for both request and response bodies.
func CallHTTP(address string, doc model.Document) (model.Document, error) {
	var out model.Document
	b, err := json.Marshal(doc)
	if err != nil {
		return out, err
	}
	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest(http.MethodPost, address, bytes.NewReader(b))
	if err != nil {
		return out, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return out, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return out, err
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return out, err
	}
	return out, nil
}
