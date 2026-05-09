package department

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/godispatcher/dispatcher/model"
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
			rw = WriteErrorDoc(errors.New("multipart form handler error:"+err.Error()), w)
			return rw
		}
		document = docTmp
	} else {
		rw = WriteErrorDoc(errors.New("bad content type"), w)
		return rw
	}

	if document.Department == "" && document.Transaction == "" {
		rw = WriteErrorDoc(errors.New("department and transaction parameters is empty"), w)
		return rw
	}

	// Store RequestContext in a goroutine-local store
	ctx := &model.RequestContext{
		Header:      r.Header,
		QueryParams: r.URL.Query(),
		URLParams:   make(map[string]string),
	}

	path := r.URL.Path
	segments := strings.Split(strings.Trim(path, "/"), "/")
	if len(segments) >= 2 {
		ctx.URLParams["department"] = segments[0]
		ctx.URLParams["transaction"] = segments[1]

		for i := 2; i < len(segments); i++ {
			ctx.URLParams[fmt.Sprintf("segment_%d", i)] = segments[i]
		}

		if document.Department == "" {
			document.Department = segments[0]
		}
		if document.Transaction == "" {
			document.Transaction = segments[1]
		}
	}
	model.SetRequestContext(ctx)
	defer model.ClearRequestContext()

	// If document.Security.VerifyCode is empty, try to obtain it from X-Verify-Code header
	if document.Security == nil || strings.TrimSpace(document.Security.VerifyCode) == "" {
		if vcode := strings.TrimSpace(r.Header.Get("X-Verify-Code")); vcode != "" {
			if document.Security == nil {
				document.Security = &model.Security{}
			}
			document.Security.VerifyCode = vcode
		}
	}
	ta := DispatcherHolder.GetTransaction(document.Department, document.Transaction)
	if ta != nil {
		outputDoc := (*ta).GetTransaction().Init(document)

		response, err := json.Marshal(outputDoc)
		if err != nil {
			rw = WriteErrorDoc(errors.New("bad transaction response"), w)
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
	outputDoc := model.Document{Error: err.Error(), Type: "Error"}
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
