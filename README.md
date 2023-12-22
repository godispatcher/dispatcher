# Dispatcher

İster küçük/orta ölçekli, ister büyük ölçekli bir hizmet tasarlıyor olun, bunların hepsini dağıtıcı mimarisiyle oluşturabilirsiniz.

Whether you are designing a small/medium-sized or large-scale service, you can create them all with a dispatcher architecture.

Tek bir function ile servisinizi sunmaya başlayın.
---
**Start serving your service with just a single function.**

src/department/example.go
```go
package department

import "github.com/denizakturk/dispatcher/middleware"

type Example stryct {
    middleware.Middleware[struct{Name string `json:"name"`}, string]
}

func (t Example) Transact()(interface{}, error){
    return "Hello "+t.Response.Name
}
```
Oluşturduğumuz işlemi dispatcher hizmetine kayıt ediyoruz.
---
**We register the operation we created with the dispatcher service.**

src/department/register.go
```go
package department

import "github.com/denizakturk/dispatcher/middleware"

func Register() {
    exampleDepartment := middleware.NewDepartmentManager()
    err := exampleDepartment.AddTransaction(middleware.NewTransactionInit("Example", "example", &Example{}))
    if err != nil {
        panic(err)
    }
    
    exampleDepartment.Register()
}
```
Son yapmamız gereken ana fonksiyona tanımlama yapmak
---
**The last thing we need to do is define the main function.**

main.go
```go
package main

import (
	"github.com/denizakturk/dispatcher/registrant"
	"github.com/denizakturk/dispatcher/server"
	"github.com/denizakturk/dispatcher/src/department"
)

func main() {
	dispatcher := registrant.NewRegisterDispatch()
	dispatcher.Port = "9001" // Default 9000 optinal
	department.Register()
	server.InitServer(dispatcher)
}
```

Port tanımlaması opsiyonel olup öntanımlı 9000 dir.

Bunları bilmelisiniz
===
**You should know these**

Güvenlik için yapacağınız jwt ve benzeri token kontrollerini veya ihtiyacınız olan yapıları kendi oluşturacağınız middleware'a dispatcher'ın sunduğu basit middelware'ı embed ederek kullanabilir veya middleware.MiddlewareInterface yardımıyla baştan kendi middleware'ınızı oluşturabilir siniz.

Esnekliği nasıl sağlayacağınıza dair örnek bir uygulama yapalım
----
**Let's create an example application for how to provide flexibility**

src/department/example.go
```go
package department

import (
    "github.com/denizakturk/dispatcher/middleware"
    "github.com/denizakturk/dispatcher/model"
)

type LicenceCheckerMiddleware[Req, Res any] struct {
	middleware.Middleware[Req, Res]
	Token string
}

func (m *LicenceCheckerMiddleware[Req, Res]) SetToken(token string) {
	m.Token = token
}
func (m LicenceCheckerMiddleware[Req, Res]) LicenceChecker(licence string) bool {
	return true
}

func (m *LicenceCheckerMiddleware[Req, Res]) InitTransaction(document model.Document) (err error) {
    err = m.Middleware.InitTransaction(document)
    if err != nil {
        return err
    }
    m.SetToken(document.Security.Licence)
    return err
}

type Example struct {
LicenceCheckerMiddleware[Request, Response]
}

func (t Example) Transact() (interface{}, error) {
response := Response{Name: t.Request.ID + " Name", Token: t.Token}

return response, nil
}

```
Dispatcher'a register ettiğiniz hizmetleri girdi/çıktı parametrelerini görmek için
---
Bir nevi dökümantasyon

Projeyi 9000 portunda local'inizde çalıştırdığınızı varsayıyoruz
    http://localhost:9000/help
adresine yapacağınız istekde size tüm department/transaction listesini verecektir, 
ayrıca kısa versiyonunu görmek için
    http://localhost:9000/help?short=1