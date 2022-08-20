package department

import (
	"dispatcher/converter"
	"dispatcher/matcher"
)

type ProductCreateRequest struct {
	Name        string `json:"name" require:"true" isEmpty:"false"`
	Slug        string `json:"slug" require:"true"`
	Description string `json:"description"`
	Counter     int    `json:"counter" require:"true" isEmpty:"false"`
}

type ProductCreateResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
}

type ProductCreate struct {
	Request  ProductCreateRequest
	Response ProductCreateResponse
	converter.TransactionExchangeConverter
}

func (t *ProductCreate) SetRequest(req string) {
	t.HandleRequest(req, &t.Request)
}

func (t *ProductCreate) GetResponse() interface{} {
	return t.Response

}

func (t *ProductCreate) GetRequestType() interface{} {
	return t.Request
}

func (t *ProductCreate) Transact() {
	t.Response.Name = t.Request.Name
	t.Response.Slug = t.Request.Slug
	t.Response.Description = t.Request.Description
	t.Response.ID = "UUID()"
}

func NewProductDepartment() {
	productDepartment := matcher.Department{}
	productDepartment.Transactions = make(map[string]matcher.Transaction)
	createTransaction := ProductCreate{}
	getByIdTransaction := ProductGetById{}
	productDepartment.Name = "Product"
	productDepartment.Slug = "product"
	productDepartment.Transactions["create"] = &createTransaction
	productDepartment.Transactions["getById"] = &getByIdTransaction

	matcher.RegisterDepartment(productDepartment)
}

type ProductGetById struct {
	Request  ProductGetByIdRequest
	Response ProductGetByIdResponse
	converter.TransactionExchangeConverter
}

func (t *ProductGetById) SetRequest(req string) {
	t.HandleRequest(req, &t.Request)
}

func (t *ProductGetById) GetResponse() interface{} {
	return t.Response

}

func (t *ProductGetById) GetRequestType() interface{} {
	return t.Request
}

type ProductGetByIdRequest struct {
	ID string `json:"id" require:"true"`
}

type ProductGetByIdResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
}

func (t *ProductGetById) Transact() {
	t.Response.Name = "Pantelon"
	t.Response.Slug = "pantelon"
	t.Response.Description = "Koyu renkli pantelon"
	t.Response.ID = "UUID()"
}
