package creator

import (
	"github.com/denizakturk/dispatcher/department"
	"github.com/denizakturk/dispatcher/middleware"
	"github.com/denizakturk/dispatcher/model"
	"github.com/denizakturk/dispatcher/server"
	"github.com/denizakturk/dispatcher/transaction"
	"net/http"
)

func NewTransaction[T any, TI transaction.Transaction[T]](departmentName, transactionName string, runables []middleware.MiddlewareRunable, options ...map[string]string) {
	tmp := transaction.TransactionBucketItem{}
	tmp.Name = transactionName
	header := http.Header{}
	if options != nil {
		for _, option := range options {
			for key, val := range option {
				header.Set(key, val)
			}
		}
	}

	tmp.Transaction = server.Server[T, TI]{Runables: runables, Options: model.ServerOption{Header: header}}

	department.DispatcherHolder.Add(departmentName, tmp)
}
