package creator

import (
	"net/http"

	"github.com/godispatcher/dispatcher/department"
	"github.com/godispatcher/dispatcher/middleware"
	"github.com/godispatcher/dispatcher/model"
	"github.com/godispatcher/dispatcher/server"
	"github.com/godispatcher/dispatcher/transaction"
)

func NewTransaction[T any, TI transaction.Transaction[T]](departmentName, transactionName string, runables []middleware.MiddlewareRunable, options ...any) {
	tmp := transaction.TransactionBucketItem{}
	tmp.Name = transactionName
	header := http.Header{}
	var transactionOptions model.TransactionOptions
	if options != nil {
		for _, option := range options {
			switch opt := option.(type) {
			case map[string]string:
				for key, val := range opt {
					header.Set(key, val)
				}
			case model.TransactionOptions:
				transactionOptions = opt
			}
		}
	}

	tmp.Transaction = server.Server[T, TI]{Runables: runables, Options: model.ServerOption{Header: header, TransactionOptions: transactionOptions}}

	department.DispatcherHolder.Add(departmentName, tmp)
}
