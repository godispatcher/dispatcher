package server

import (
	"encoding/json"
	"fmt"
	"github.com/godispatcher/dispatcher/constants"
	"github.com/godispatcher/dispatcher/department"
	"github.com/godispatcher/dispatcher/middleware"
	"github.com/godispatcher/dispatcher/model"
	"github.com/godispatcher/dispatcher/transaction"
	"github.com/godispatcher/dispatcher/utilities"
	"log"
	"net/http"
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

func ServJsonApi(register *department.RegisterDispatcher) {
	http.Handle("/", register)
	log.Fatal(http.ListenAndServe(":"+register.Port, nil))
}
