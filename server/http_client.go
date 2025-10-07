package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
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
	// Normalize URL: ensure it has a trailing slash if no path is provided
	u, err := url.Parse(address)
	if err == nil {
		if u.Path == "" {
			u.Path = "/"
			address = u.String()
		}
	}
	client := &http.Client{Timeout: 15 * time.Second}
	mkReq := func(closeConn bool) (*http.Request, error) {
		req, err := http.NewRequest(http.MethodPost, address, bytes.NewReader(b))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		// Propagate verify code if present in the document security
		if doc.Security != nil && strings.TrimSpace(doc.Security.VerifyCode) != "" {
			req.Header.Set("X-Verify-Code", doc.Security.VerifyCode)
		}
		if closeConn {
			req.Header.Set("Connection", "close")
		}
		return req, nil
	}
	// First attempt
	req, err := mkReq(false)
	if err != nil {
		return out, err
	}
	resp, err := client.Do(req)
	if err != nil {
		// Retry once on EOF or connection closed errors, forcing Connection: close
		if errors.Is(err, io.EOF) || strings.Contains(err.Error(), "EOF") || strings.Contains(strings.ToLower(err.Error()), "use of closed network connection") {
			req2, err2 := mkReq(true)
			if err2 != nil {
				return out, err
			}
			resp, err = client.Do(req2)
		}
	}
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
