package department

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/godispatcher/dispatcher/model"
	"io"
	"net/http"
	"strconv"
	"strings"
)

const (
	ContentTypeJSON           = "application/json"
	ContentTypeXML            = "application/xml"
	ContentTypeFormURLEncoded = "application/x-www-form-urlencoded"
	ContentTypeMultipart      = "multipart/form-data"
	ContentTypePlainText      = "text/plain"
	ContentTypeHTML           = "text/html"
)

func RegisterMainFunc(w http.ResponseWriter, r *http.Request) (rw model.RegisterResponseModel) {
	var document model.Document
	ct := r.Header.Get("Content-Type")
	if strings.HasPrefix(ct, ContentTypeJSON) {
		docTmp, err := JsonHandler(r)
		if err != nil {
			rw = WriteErrorDoc(err, w)
			return rw
		}
		document = docTmp
	} else if strings.HasPrefix(ct, ContentTypeFormURLEncoded) {
		docTmp, err := UrlEncodedHandler(r)
		if err != nil {
			rw = WriteErrorDoc(err, w)
			return rw
		}
		document = docTmp
	} else if strings.HasPrefix(ct, ContentTypeMultipart) {
		docTmp, err := MultipartFormHandler(r)
		if err != nil {
			rw = WriteErrorDoc(errors.New("dad content type"), w)
			return rw
		}
		document = docTmp
	} else {
		rw = WriteErrorDoc(errors.New("dad content type"), w)
		return rw
	}

	if &document == nil {
		rw = WriteErrorDoc(errors.New("dad content type"), w)
		return rw
	}
	ta := DispatcherHolder.GetTransaction(document.Department, document.Transaction)
	if ta != nil {
		outputDoc := (*ta).GetTransaction().Init(document)

		response, err := json.Marshal(outputDoc)
		if err != nil {
			rw = WriteErrorDoc(errors.New("dad content type"), w)
			return rw
		}

		if document.Dispatchings != nil {
			for _, v := range document.Dispatchings {
				cta := DispatcherHolder.GetTransaction(v.Department, v.Transaction)
				if cta != nil {
					dOutputDoc := (*cta).GetTransaction().Init(*v)
					outputDoc.Dispatchings = append(outputDoc.Dispatchings, &dOutputDoc)
					//TODO: if Ignored errors add dispatching ignoredError option params in model.document
					if dOutputDoc.Error != nil {
						break
					}
				}
			}
		}
		options := (*ta).GetTransaction().GetOptions()
		if &options != nil {
			for key, _ := range options.Header {
				w.Header().Set(key, options.Header.Get(key))
			}
		}
		rw.Header = w.Header()
		rw.Body = string(response)
		fmt.Fprint(w, string(response))
		return rw
	}
	outputDoc := model.Document{Department: document.Department, Transaction: document.Transaction, Error: errors.New("transaction not found").Error(), Type: "Error"}
	w.WriteHeader(http.StatusBadRequest)
	rw.StatusCode = http.StatusBadRequest
	response, err := json.Marshal(outputDoc)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		rw.StatusCode = http.StatusBadRequest
		return rw
	}
	rw.Body = string(response)
	fmt.Fprint(w, string(response))
	return rw
}

func WriteErrorDoc(err error, w http.ResponseWriter) (rw model.RegisterResponseModel) {
	outputDoc := model.Document{Error: errors.New("transaction not found").Error(), Type: "Error"}
	w.WriteHeader(http.StatusBadRequest)
	rw.StatusCode = http.StatusBadRequest
	response, err := json.Marshal(outputDoc)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		rw.StatusCode = http.StatusBadRequest
		return rw
	}
	rw.Body = response
	fmt.Fprint(w, string(response))

	return rw
}

func JsonHandler(r *http.Request) (model.Document, error) {
	document := model.Document{}
	bodyByte, err := io.ReadAll(r.Body)
	if err != nil {
		return document, err
	}
	err = json.Unmarshal(bodyByte, &document)
	return document, err
}

func UrlEncodedHandler(r *http.Request) (model.Document, error) {
	document := model.Document{}
	err := r.ParseForm()
	if err != nil {
		return document, err
	}
	form := ConvertSliceAtoi(r.Form)
	byteJson, err := json.Marshal(form)

	if err != nil {
		return document, err
	}
	path := r.URL.Path

	segments := strings.Split(strings.Trim(path, "/"), "/")
	if len(segments) < 2 {
		return document, errors.New("invalid path")
	}
	document.Department = segments[0]
	document.Transaction = segments[1]
	err = json.Unmarshal(byteJson, &document.Form)
	return document, err
}
func MultipartFormHandler(r *http.Request) (model.Document, error) {
	document := model.Document{}
	r.ParseMultipartForm(32 << 20)
	form := ConvertSliceAtoi(r.MultipartForm.Value)
	byteJson, err := json.Marshal(form)
	if err != nil {
		return document, err
	}
	path := r.URL.Path

	segments := strings.Split(strings.Trim(path, "/"), "/")
	if len(segments) < 2 {
		return document, errors.New("invalid path")
	}
	document.Department = segments[0]
	document.Transaction = segments[1]
	err = json.Unmarshal(byteJson, &document.Form)
	return document, err
}

func ConvertSliceAtoi(slice map[string][]string) map[string]any {
	result := make(map[string]any, len(slice))
	for key, val := range slice {

		var res []any
		for _, v := range val {
			var convV any
			if conv, err := strconv.Atoi(v); err == nil {
				convV = conv
			} else {
				convV = val[0]
			}
			res = append(res, convV)
		}
		if len(res) < 2 {
			result[key] = res[0]
		} else {
			result[key] = res
		}
	}

	return result
}
