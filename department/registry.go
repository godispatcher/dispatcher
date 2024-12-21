package department

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/denizakturk/dispatcher/model"
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

func RegisterMainFunc(w http.ResponseWriter, r *http.Request) {
	var document model.Document
	if r.Header.Get("Content-Type") == ContentTypeJSON {
		docTmp, err := JsonHandler(r)
		if err != nil {
			WriteErrorDoc(err, w)
			return
		}
		document = docTmp
	} else if r.Header.Get("Content-Type") == ContentTypeFormURLEncoded {
		docTmp, err := UrlEncodedHandler(r)
		if err != nil {
			WriteErrorDoc(err, w)
			return
		}
		document = docTmp
	} else if strings.HasPrefix(r.Header.Get("Content-Type"), ContentTypeMultipart) {
		docTmp, err := MultipartFormHandler(r)
		if err != nil {
			WriteErrorDoc(errors.New("dad content type"), w)
			return
		}
		document = docTmp
	} else {
		WriteErrorDoc(errors.New("dad content type"), w)
		return
	}

	if &document == nil {
		WriteErrorDoc(errors.New("dad content type"), w)
		return
	}
	ta := DispatcherHolder.GetTransaction(document.Department, document.Transaction)
	if ta != nil {
		outputDoc := (*ta).GetTransaction().Init(document)

		response, err := json.Marshal(outputDoc)
		if err != nil {
			WriteErrorDoc(errors.New("dad content type"), w)
			return
		}

		if document.Dispatchings != nil {
			for _, v := range document.Dispatchings {
				cta := DispatcherHolder.GetTransaction(v.Department, v.Transaction)
				if cta != nil {
					dOutputDoc := (*cta).GetTransaction().Init(*v)
					outputDoc.Dispatchings = append(outputDoc.Dispatchings, &dOutputDoc)
					// if Ignored errors add dispatching ignoredError option params in model.document
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
		fmt.Fprint(w, string(response))
		return
	}
	outputDoc := model.Document{Department: document.Department, Transaction: document.Transaction, Error: errors.New("transaction not found")}
	w.WriteHeader(http.StatusBadRequest)
	response, err := json.Marshal(outputDoc)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	fmt.Fprint(w, string(response))
	return
}

func WriteErrorDoc(err error, w http.ResponseWriter) {
	outputDoc := model.Document{Error: errors.New("transaction not found")}
	w.WriteHeader(http.StatusBadRequest)
	response, err := json.Marshal(outputDoc)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	fmt.Fprint(w, string(response))
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
