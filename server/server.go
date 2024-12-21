package server

import (
	"encoding/json"
	"fmt"
	"github.com/denizakturk/dispatcher/constants"
	"github.com/denizakturk/dispatcher/department"
	"github.com/denizakturk/dispatcher/model"
	"github.com/denizakturk/dispatcher/transaction"
	"github.com/denizakturk/dispatcher/utilities"
	"log"
	"net/http"
)

type Server[T any, TI transaction.Transaction[T]] struct {
	Options model.ServerOption
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

func (Server[T, TI]) Init(document model.Document) model.Document {
	var ta TI = new(T)
	jsonByteData, err := json.Marshal(document.Form)
	if err != nil {
		outputErrDoc := model.Document{Department: document.Department, Transaction: document.Transaction, Error: err}
		return outputErrDoc
	}

	if err != nil {
		outputErrDoc := model.Document{Department: document.Department, Transaction: document.Transaction, Error: err}
		return outputErrDoc
	}
	validator := model.DocumentFormValidater{Request: string(jsonByteData)}
	err = validator.Validate(ta.GetRequest())
	if err != nil {
		outputErrDoc := model.Document{Department: document.Department, Transaction: document.Transaction, Error: err.Error()}
		return outputErrDoc
	}
	err = ta.Transact()
	if err != nil {
		outputErrDoc := model.Document{Department: document.Department, Transaction: document.Transaction, Error: err}
		return outputErrDoc
	}
	ta.SetRequest(jsonByteData)
	err = ta.Transact()
	if err != nil {
		outputErrDoc := model.Document{Department: document.Department, Transaction: document.Transaction, Error: err}
		return outputErrDoc
	}
	document.Output = ta.GetResponse()

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
			transaction.Name = v.GetName()
			if !r.URL.Query().Has("short") || r.URL.Query().Get("short") == "0" {
				nestedTypeCtrl = &[]string{}
				transaction.Procudure = utilities.Analysis(v.GetTransaction().GetRequest(), nestedTypeCtrl)
				nestedTypeCtrl = &[]string{}
				transaction.Output = utilities.Analysis(v.GetTransaction().GetResponse(), nestedTypeCtrl)
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
