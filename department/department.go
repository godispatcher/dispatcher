package department

import (
	"github.com/denizakturk/dispatcher/transaction"
	"net/http"
)

type Department struct {
	Name         string
	Transactions []transaction.TransactionBucketItemInterface
}

type DispacherBucket []Department

func (db *DispacherBucket) Add(name string, transaction transaction.TransactionBucketItemInterface) {
	hasDepartment := false
	hasTransaction := false
	for _, val := range *db {
		if val.Name == name {
			hasDepartment = true
			for _, v := range val.Transactions {
				for v.GetName() == transaction.GetName() {
					hasTransaction = true
					val.Transactions = append(val.Transactions, transaction)
				}
			}
		}
	}
	if !hasDepartment && !hasTransaction {
		tmpDep := Department{}
		tmpDep.Name = name
		tmpDep.Transactions = append(tmpDep.Transactions, transaction)
		*db = append(*db, tmpDep)
	}
}

func (db *DispacherBucket) GetTransaction(departmentName, transactionName string) *transaction.TransactionBucketItemInterface {
	for _, val := range *db {
		if val.Name == departmentName {
			for _, v := range val.Transactions {
				for v.GetName() == transactionName {
					return &v
				}
			}
		}
	}

	return nil
}

var DispatcherHolder DispacherBucket

func NewRegisteryDispatcher(port string) *RegisterDispatcher {
	return &RegisterDispatcher{Port: port, MainFunc: RegisterMainFunc}
}

type RegisterDispatcher struct {
	MainFunc func(http.ResponseWriter, *http.Request)
	Port     string
}

func (rd RegisterDispatcher) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rd.MainFunc(w, r)
}
