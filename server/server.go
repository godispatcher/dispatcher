package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/godispatcher/dispatcher/constants"
	"github.com/godispatcher/dispatcher/department"
	"github.com/godispatcher/dispatcher/middleware"
	"github.com/godispatcher/dispatcher/model"
	"github.com/godispatcher/dispatcher/transaction"
	"github.com/godispatcher/dispatcher/utilities"
)

type Server[T any, TI transaction.Transaction[T]] struct {
	Options  model.ServerOption
	Runables []middleware.MiddlewareRunable
}

func (s *Server[T, TI]) AddRunable(runable middleware.MiddlewareRunable) {
	s.Runables = append(s.Runables, runable)
}

func (Server[T, TI]) GetRequest() any {
	var ta TI = new(T)
	return ta.GetRequest()
}
func (Server[T, TI]) GetResponse() any {
	var ta TI = new(T)
	return ta.GetResponse()
}

func (s Server[T, TI]) GetOptions() model.ServerOption {
	return s.Options
}

func (s Server[T, TI]) Init(document model.Document) model.Document {
	var ta TI = new(T)
	if s.Runables != nil {
		ta.SetRunables(s.Runables)
	}

	err := ta.SetSelfRunables()

	if err != nil {
		outputErrDoc := model.Document{Department: document.Department, Transaction: document.Transaction, Error: err.Error(), Type: "Error"}
		return outputErrDoc
	}

	jsonByteData, err := json.Marshal(document.Form)
	if err != nil {
		outputErrDoc := model.Document{Department: document.Department, Transaction: document.Transaction, Error: err.Error(), Type: "Error"}
		return outputErrDoc
	}

	if err != nil {
		outputErrDoc := model.Document{Department: document.Department, Transaction: document.Transaction, Error: err.Error(), Type: "Error"}
		return outputErrDoc
	}
	validator := model.DocumentFormValidater{Request: string(jsonByteData)}
	err = validator.Validate(ta.GetRequest())
	if err != nil {
		outputErrDoc := model.Document{Department: document.Department, Transaction: document.Transaction, Error: err.Error(), Type: "Error"}
		return outputErrDoc
	}
	if ta.GetRunables() != nil {
		for _, runF := range ta.GetRunables() {
			err := runF(document)
			if err != nil {
				outputErrDoc := model.Document{Department: document.Department, Transaction: document.Transaction, Error: err.Error(), Type: "Error"}
				return outputErrDoc
			}
		}
	}
	ta.SetRequest(jsonByteData)
	err = ta.Transact()
	if err != nil {
		outputErrDoc := model.Document{Department: document.Department, Transaction: document.Transaction, Error: err.Error(), Type: "Error"}
		return outputErrDoc
	}
	document.Output = ta.GetResponse()
	document.Type = "Result"

	return document
}

type TransactionListHelper struct {
	Name      string      `json:"name"`
	Procudure interface{} `json:"procedure,omitempty"`
	Output    interface{} `json:"output,omitempty"`
}
type DepartmentListHelper struct {
	Name         string                  `json:"name"`
	Transactions []TransactionListHelper `json:"transactions"`
}

type HelperList struct {
	Departments []DepartmentListHelper `json:"departments"`
}

type ApiDocServer struct {
}

func (ApiDocServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	helperList := HelperList{}
	var nestedTypeCtrl *[]string
	for _, val := range department.DispatcherHolder {
		department := DepartmentListHelper{}
		department.Name = val.Name

		for _, v := range val.Transactions {
			transaction := TransactionListHelper{}
			transaction.Name = (*v).GetName()
			if !r.URL.Query().Has("short") || r.URL.Query().Get("short") == "0" {
				nestedTypeCtrl = &[]string{}
				transaction.Procudure = utilities.Analysis((*v).GetTransaction().GetRequest(), nestedTypeCtrl)
				nestedTypeCtrl = &[]string{}
				transaction.Output = utilities.Analysis((*v).GetTransaction().GetResponse(), nestedTypeCtrl)
			}
			department.Transactions = append(department.Transactions, transaction)
		}
		helperList.Departments = append(helperList.Departments, department)
	}
	response, _ := json.Marshal(helperList)
	w.Header().Add(constants.HTTP_CONTENT_TYPE, constants.HTTP_CONTENT_JSON)
	fmt.Fprint(w, string(response))
}

func ServJsonApiDoc() {
	http.Handle("/help", ApiDocServer{})
}

// ServJsonApi starts the HTTP server and applies CORS/same-origin controls if configured
func ServJsonApi(register *department.RegisterDispatcher) {
	var handler http.Handler = register
	if register != nil && register.CORS != nil {
		handler = withCORS(handler, register.CORS)
	} else {
		// apply sensible defaults (permissive CORS) to allow external control later
		defaults := (&model.CORSOptions{}).WithDefaults()
		handler = withCORS(handler, defaults)
	}
	http.Handle("/", handler)
	log.Fatal(http.ListenAndServe(":"+register.Port, nil))
}

// withCORS wraps the given handler with CORS and optional same-origin enforcement
func withCORS(next http.Handler, opts *model.CORSOptions) http.Handler {
	options := opts.WithDefaults()
	allowedMethods := strings.Join(options.AllowedMethods, ", ")
	allowedHeaders := strings.Join(options.AllowedHeaders, ", ")
	exposeHeaders := strings.Join(options.ExposeHeaders, ", ")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if options.EnforceSameOrigin && origin != "" {
			if !sameOrigin(origin, r.Host) {
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte("forbidden: same-origin policy enforced"))
				return
			}
		}

		// Preflight handling
		if r.Method == http.MethodOptions {
			applyCORSHeaders(w, origin, options, allowedMethods, allowedHeaders, exposeHeaders, r.Header.Get("Access-Control-Request-Headers"))
			// 204 No Content for preflight
			w.WriteHeader(http.StatusNoContent)
			return
		}

		applyCORSHeaders(w, origin, options, allowedMethods, allowedHeaders, exposeHeaders, r.Header.Get("Access-Control-Request-Headers"))
		next.ServeHTTP(w, r)
	})
}

func applyCORSHeaders(w http.ResponseWriter, origin string, options *model.CORSOptions, allowedMethods, allowedHeaders, exposeHeaders, reqHeaders string) {
	w.Header().Add("Vary", "Origin")
	if origin != "" {
		if originAllowed(origin, options.AllowedOrigins) {
			if options.AllowCredentials {
				// When credentials are allowed, must echo specific origin instead of '*'
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			} else {
				// Allow all or specific
				if oneStar(options.AllowedOrigins) {
					w.Header().Set("Access-Control-Allow-Origin", "*")
				} else {
					w.Header().Set("Access-Control-Allow-Origin", origin)
				}
			}
		}
		if exposeHeaders != "" {
			w.Header().Set("Access-Control-Expose-Headers", exposeHeaders)
		}
		w.Header().Set("Access-Control-Allow-Methods", allowedMethods)
		if reqHeaders != "" {
			w.Header().Set("Access-Control-Allow-Headers", reqHeaders)
		} else if allowedHeaders != "" {
			w.Header().Set("Access-Control-Allow-Headers", allowedHeaders)
		}
		if options.MaxAge > 0 {
			w.Header().Set("Access-Control-Max-Age", fmt.Sprintf("%d", options.MaxAge))
		}
	}
}

func oneStar(allowed []string) bool {
	return len(allowed) == 1 && allowed[0] == "*"
}

func originAllowed(origin string, allowed []string) bool {
	if oneStar(allowed) {
		return true
	}
	for _, a := range allowed {
		if strings.EqualFold(a, origin) {
			return true
		}
	}
	return false
}

func sameOrigin(origin, host string) bool {
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	// Compare host:port; scheme is generally irrelevant for same-origin in simple check
	return strings.EqualFold(u.Host, host)
}
