package handling

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/denizakturk/dispatcher/constants"
	"github.com/denizakturk/dispatcher/model"
)

const ContentTypeApplicationJson = "application/json"

func restToDispatcher(req *http.Request) ([]byte, error) {
	route := strings.Split(strings.Trim(req.URL.Path, "/"), "/")
	if len(route) != 2 {
		return nil, fmt.Errorf("url path does not have 2 part")
	}
	document := model.Document{}
	document.Department = route[0]
	document.Transaction = route[1]
	if strings.Contains(req.Header.Get("Authorization"), "Bearer ") {
		document.Security = &model.Security{}
		document.Security.Licence = strings.Split(req.Header.Get("Authorization"), "Bearer ")[1]
	}

	for key, _ := range req.Form {
		if document.Form == nil {
			document.Form = make(map[string]interface{})
		}
		a, err := strconv.ParseInt(req.FormValue(key)[0:], 10, 64)
		if err != nil {
			document.Form[key] = req.FormValue(key)
		} else {
			document.Form[key] = a
		}

	}
	return json.Marshal(document)
}

func RequestHandle(req *http.Request) ([]byte, error) {
	if strings.Contains(req.Header.Get("Content-Type"), "application/x-www-form-urlencoded") {
		err := req.ParseForm()
		if err != nil {
			return nil, err
		}
		return restToDispatcher(req)
	}

	if strings.Contains(req.Header.Get("Content-Type"), "multipart/form-data") {
		err := req.ParseMultipartForm(0)
		if err != nil {
			return nil, err
		}
		return restToDispatcher(req)
	}

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
