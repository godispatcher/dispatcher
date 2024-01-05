package middleware

import (
	"github.com/denizakturk/dispatcher/model"
	"github.com/denizakturk/dispatcher/registrant"
	"strings"
)

func NewDepartmentManager() DepartmentsManagement {
	dm := DepartmentsManagement{}
	dm.Departments = make(map[string]*model.Department)

	return dm
}

type DepartmentsManagement struct {
	Departments map[string]*model.Department
}

func (m *DepartmentsManagement) AddTransaction(transaction TransactionInit) error {
	if _, ok := m.Departments[transaction.Department]; !ok {
		m.Departments[transaction.Department] = &model.Department{
			Name: transaction.Department,
			Slug: strings.ToLower(transaction.Department),
		}
	}
	transaction.Type.Defaults()
	if _, ok := m.Departments[transaction.Department]; ok {
		tmp := model.TransactionHolder{
			Name:            transaction.Transaction,
			Type:            transaction.Type,
			InitTransaction: transaction.Init,
			Options: model.TransactionOptions{
				Security: model.SecurityOptions{
					LicenceChecker: transaction.Type.IsLicenceRequired(),
				},
			},
			LicenceValidator: transaction.Type.LicenceChecker,
		}

		m.Departments[transaction.Department].Transactions = append(m.Departments[transaction.Department].Transactions, tmp)
	}

	return nil
}

func (m *DepartmentsManagement) Register() {
	for _, val := range m.Departments {
		registrant.RegisterDepartment(*val)
	}
}
