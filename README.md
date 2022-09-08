# Dispatcher Macroservice - Dispatcher Macro Hizmeti

## Hizmet oluşturmak için yapmanız gereken bir Department type'ını kullanarak değişken oluşturmak, Departmant içindeki Transactions altına her bir transaction için bir struct tanımlayıp Transaction interface'ine uymaktır.
---
## To create a service, all you need to do is create a department struct and define a struct for each transaction and to apply the Transaction interface.


```go

import (
	"github.com/denizakturk/dispatcher/handling"
	"github.com/denizakturk/dispatcher/model"
	"github.com/denizakturk/dispatcher/registrant"
)

type Department struct {
	Name         string `json:"name"`
	Slug         string `json:"slug"`
	Transactions map[string]Transaction
}

type Transaction interface {
	Transact() error
	SetRequest(req string)
	GetRequest() interface{}
	GetResponse() interface{}
	GetOptions() TransactionOptions
	LicenceChecker(licence string) bool
}
```
## **Example department and transaction infrastruct**

```go

import

// Request - Response structs
type ProductCreateRequest struct {
    Name string `json:"name"`
    Slug string `json:"slug"`
    Description string `json:"description"`
}
type ProductCreateResponse struct {
    ID int `json:"id"`
    Name string `json:"name"`
    Slug string `json:"slug"`
    Description string `json:"description"`
}

// Product create transaction
type ProductCreate struct {
    Request ProductCreateRequest
    Response ProductCreateResponse
    handling.TransactionExchangeConverter // Helper struct
	model.TransactionOptions // Options
}


func (t *ProductCreate) SetRequest(req string) {
	t.HandleRequest(req, &t.Request)
}

func (t *ProductCreate) GetResponse() interface{} {
	return t.Response

}

func (t *ProductCreate) GetRequest() interface{} {
	return t.Request
}

func (t *ProductCreate) Transact() {
	t.Response.Name = "Product 1"
	t.Response.Slug = "product-1"
	t.Response.Description = "Product 1 Description"
	t.Response.ID = 12314
}

productDepartment := model.Department{Name:"Product", Slug:"product"}
productCreate := &ProductCreate{}
productCreate.Security.LicenceChecker = true
productDepartment.Transactions["create"] = productCreate

registrant.RegisterDepartment(productDepartment)

```

### Service Requesting

#### Bir transaction yürütmek için Document tipinde bir json oluşturup bunu http isteği olara header kısmında "Content-Type:application(json" olarak belirtmelisiniz.

```go
type Document struct {
	Department         string              `json:"department,omitempty"`
	Transaction        string              `json:"transaction,omitempty"`
	Type               string              `json:"type,omitempty"`
	Procedure          interface{}         `json:"procedure,omitempty"`
	Form               DocumentForm        `json:"form,omitempty"`
	Output             interface{}         `json:"output,omitempty"`
	Error              interface{}         `json:"error,omitempty"`
	Dispatchings       []*Document         `json:"dispatchings,omitempty"`
	ChainRequestOption ChainRequestOption  `json:"chain_request_option,omitempty"`
	Security           *Security           `json:"security,omitempty"`
	Options            *TransactionOptions `json:"options,omitempty"`
}
```

## **Department**:
#### İlgili transaction'ın barındığı departman ismi, bunu oluştururken veriyorsunuz.
---
## **Transaction**:
#### İlgili işlemin ismi, aynı işlem isimlerini farklı department'lar altında tanımlayabilirsiniz.
---
## **Type**:
#### Servis çağrılarında bu alan boş bırakılabilir, procedure yazıldığında o transaction'ın aldığı parametreler listelenecektir.
---
## **Procedure**:
#### Type=procedure olarak işaretlendiğinde bu alanın içine ilgili tranaction parametreleri gelecektir.
---
## **Form**:
#### Transaction için ihtiyaç duyulan parametreler form alanında gönderilir.
---
## **Output**:
#### Transaction çıktısı buraya basılır.
---
## **Error**:
#### Herhangi bir hata durumunda buraya hata mesajı basılır.
---
## **Dispatchings**:
#### Bu kısımda bir sonraki yapılacak işlem tanımlanır ve ChainRequestOption ile desteklenir ama bu kısım için ayrıntılı bir anlatım gerekmektedir.
---
## **ChainRequestOption**:
#### Dispatchings ile birlikte kullanılan bir alandır, Dispatchings için yapılan ek açıklama geçerlidir.
## **Security**:
#### Security altında bulunan licence değerine sisteminizde kullanılan token'ı vererek transaction'ın gerektirdiği verification admını sağlayabilirsiniz.
## **Options**:
#### Transaction'ın gerekitiği ayarları buradan kontrol edebilirsiniz, transaction'ınızı oluştururken Options->Security->LicenceChecker parametresine true vererek güvenlik adımı ekleyebilirsiniz.
## Example Document

```json
{
    "department":"Product",
    "transaction":"create",
    "form":{
		"name":"Mavi ayakkabı", 
		"slug":"mavi-ayakkabi", 
		"description":"Mavi "
	},
    "chain_request_option":{
		"id":"product_id"
	},
    "dispatchings":[
		{
			"department":"Listing", 
			"transaction":"add-product-to-list"
		}
	],
	"security":{
		"licence":"LICENCE_STRING"
	}
}
```