package server

import (
	"github.com/denizakturk/dispatcher/constants"
	"github.com/denizakturk/dispatcher/model"
	"github.com/denizakturk/dispatcher/registrant"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func InitServer(register registrant.RegisterDispatcher) {
	http.HandleFunc("/", register.MainFunc)
	http.HandleFunc("/help", RequestHelper)
	log.Fatal(http.ListenAndServe(":"+register.Port, nil))
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

func RequestHelper(res http.ResponseWriter, req *http.Request) {
	helperList := HelperList{}
	for _, val := range registrant.DepartmentRegistering {
		department := DepartmentListHelper{}
		department.Name = val.Name
		for key, v := range val.Transactions {
			transaction := TransactionListHelper{}
			transaction.Name = key
			if !req.URL.Query().Has("short") || req.URL.Query().Get("short") == "0" {
				requestProcedure := model.Procedure{}
				responseProcedure := model.Procedure{}
				requestProcedure.FromRequestType(v.GetRequestType())
				responseProcedure.FromRequestType(v.GetResponse())
				transaction.Procudure = requestProcedure
				transaction.Output = responseProcedure
			}
			department.Transactions = append(department.Transactions, transaction)
		}
		helperList.Departments = append(helperList.Departments, department)
	}
	response, _ := json.Marshal(helperList)
	res.Header().Add(constants.HTTP_CONTENT_TYPE, constants.HTTP_CONTENT_JSON)
	fmt.Fprint(res, string(response))
}
