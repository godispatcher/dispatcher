# Dispatcher Macroservice - Dispatcher Macro Hizmeti

## Hizmet oluşturmak için yapmanız gereken bir Department type'ını kullanarak değişken oluşturmak, Departmant içindeki Transactions altına her bir transaction için bir struct tanımlayıp Transaction interface'ine uymaktır.
---
## To create a service, all you need to do is create a department struct and define a struct for each transaction and to apply the Transaction interface.


```go

type Department struct {
	Name         string `json:"name"`
	Slug         string `json:"slug"`
	Transactions map[string]Transaction
}

type Transaction interface {
	Transact()
	SetRequest(req string)
	GetRequestType() interface{}
	GetResponse() interface{}
}
```
## Example department and transaction infrastruct

```go

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
	t.Response.Name = "Product 1"
	t.Response.Slug = "product-1"
	t.Response.Description = "Product 1 Description"
	t.Response.ID = 12314
}

productDepartment := model.Department{Name:"Product", Slug:"product"}

productDepartment.Transactions["create"] = &ProductCreate{}

```

### Service Requesting

#### Bir transaction yürütmek için Document tipinde bir json oluşturup bunu http isteği olara header kısmında "Content-Type:application(json" olarak belirtmelisiniz.

```go
type Document struct {
	Department         string             `json:"department,omitempty"`
	Transaction        string             `json:"transaction,omitempty"`
	Type               string             `json:"type,omitempty"`
	Procedure          interface{}        `json:"procedure,omitempty"`
	Form               DocumentForm       `json:"form,omitempty"`
	Output             interface{}        `json:"output,omitempty"`
	Error              interface{}        `json:"error,omitempty"`
	Dispatchings       []*Document         `json:"dispatchings,omitempty"`
	ChainRequestOption ChainRequestOption `json:"chain_request_option,omitempty"`
}
```

## Department:
#### İlgili transaction'ın barındığı departman ismi, bunu oluştururken veriyorsunuz.
---
## Transaction:
#### İlgili işlemin ismi, aynı işlem isimlerini farklı department'lar altında tanımlayabilirsiniz.
---
## Type:
#### Servis çağrılarında bu alan boş bırakılabilir, procedure yazıldığında o transaction'ın aldığı parametreler listelenecektir.
---
## Procedure:
#### Type=procedure olarak işaretlendiğinde bu alanın içine ilgili tranaction parametreleri gelecektir.
---
## Output:
#### Transaction çıktısı buraya basılır.
---
## Error:
#### Herhangi bir hata durumunda buraya hata mesajı basılır.
---
## Dispatchings:
#### Bu kısımda bir sonraki yapılacak işlem tanımlanır ve ChainRequestOption ile desteklenir ama bu kısım için ayrıntılı bir anlatım gerekmektedir.
---
## ChainRequestOption:
#### Dispatchings ile birlikte kullanılan bir alandır, Dispatchings için yapılan ek açıklama geçerlidir.