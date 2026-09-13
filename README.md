# GoDispatcher Framework

**GoDispatcher**, küçük ölçekli projelerden büyük ölçekli enterprise uygulamalara kadar her türlü HTTP servisini kolayca oluşturmanızı sağlayan, type-safe ve middleware tabanlı bir Go framework'üdür.

**GoDispatcher** is a type-safe, middleware-based Go framework that enables you to easily create HTTP services from small-scale projects to large-scale enterprise applications.

## 🚀 Özellikler / Features

- **Type-Safe**: Go generics kullanarak compile-time type safety
- **Middleware System**: Esnek ve genişletilebilir middleware desteği
- **Department Architecture**: Servisleri mantıksal departmanlara ayırma
- **Built-in Logging**: Otomatik request/response loglama
- **API Documentation**: Otomatik API dokümantasyonu (/help endpoint - HTML, JSON ve Toon formatları)
- **Validation System**: Struct tag'leri ile recursive ve tip güvenli request validasyonu (`require`, `is_empty`)
- **CORS Support**: Kolay yapılandırılabilir CORS desteği
- **Security**: Built-in licence validation ve güvenlik özellikleri
- **Request Chaining**: Zincirleme request desteği
- **JSON API**: RESTful JSON API desteği

## 📦 Kurulum / Installation

```bash
go mod init your-project
go get github.com/godispatcher/dispatcher
```

## ⚡ Hızlı Başlangıç / Quick Start

1) Depoyu klonlayın ve bağımlılıkları alın:

```bash
git clone https://github.com/godispatcher/dispatcher.git
cd dispatcher
go mod download
```

2) Örnek servisi çalıştırın:

```bash
go run .
```

- HTTP: http://localhost:9000
- API Dokümantasyonu: http://localhost:9000/help (HTML)
- API Dokümantasyonu (JSON): http://localhost:9000/help?format=json
- API Dokümantasyonu (Toon): http://localhost:9000/help?format=toon
- API Dokümantasyonu (Kısa): http://localhost:9000/help?short=1
- Stream (TCP/NDJSON): 9001 (HTTP portu + 1)

3) Hızlı test (curl):

```bash
curl -s http://localhost:9000/ \
  -H 'Content-Type: application/json' \
  -d '{"department":"Product","transaction":"getA","form":{}}'
```

Daha fazla bilgi ve örnekler:
- Ayrıntılı Kılavuz (TR): docs/guide-tr.md
- Advanced (EN): docs/advanced.md

## 🏗️ Temel Kullanım / Basic Usage

### 1. Transaction Oluşturma / Creating a Transaction

```go
package department

import (
    "github.com/godispatcher/dispatcher/src/middleware"
)

// Request ve Response tiplerini tanımlayın
type HelloRequest struct {
    Name string `json:"name"`
}

type HelloResponse struct {
    Message string `json:"message"`
}

// Transaction struct'ınızı oluşturun
type HelloTransaction struct {
    middleware.Middleware[HelloRequest, HelloResponse]
}

// SetSelfRunables middleware'leri ayarlamak için kullanılır
func (t *HelloTransaction) SetSelfRunables() error {
    // Burada kendi middleware'lerinizi ekleyebilirsiniz
    return nil
}

// Transact ana iş mantığınızı içerir
func (t *HelloTransaction) Transact() error {
    t.Response.Message = "Hello, " + t.Request.Name + "!"
    return nil
}
```

### 2. Transaction'ı Kaydetme / Registering Transaction

```go
package department

import "github.com/godispatcher/dispatcher/creator"

func RegisterHello() {
    creator.NewTransaction[HelloTransaction, *HelloTransaction](
        "Greeting",    // Department name
        "hello",       // Transaction name
        nil,           // Middleware runables (optional)
        nil,           // Options (optional)
    )
}
```

### 3. Ana Uygulama / Main Application

```go
package main

import (
    "github.com/godispatcher/dispatcher/department"
    "github.com/godispatcher/dispatcher/server"
    "your-project/department" // Your department package
)

func main() {
    // Dispatcher'ı başlat
    service := department.NewRegisteryDispatcher("8080")

    // Department'ları kaydet
    department.RegisterHello()

    // API dokümantasyonunu aktifleştir
    server.ServJsonApiDoc()

    // Serveri başlat
    server.ServJsonApi(service)
}
```

## 🔧 Gelişmiş Kullanım / Advanced Usage

### Custom Middleware Oluşturma / Creating Custom Middleware

```go
package department

import (
    "errors"
    "github.com/godispatcher/dispatcher/middleware"
    "github.com/godispatcher/dispatcher/model"
)

// Custom middleware struct'ı
type AuthMiddleware[Req, Res any] struct {
    middleware.Middleware[Req, Res]
    Token string
    UserID string
}

// Token doğrulama middleware'i
func (m *AuthMiddleware[Req, Res]) ValidateToken(document model.Document) error {
    if document.Security == nil || document.Security.Licence == "" {
        return errors.New("token required")
    }

    // Token doğrulama mantığınız
    if !m.isValidToken(document.Security.Licence) {
        return errors.New("invalid token")
    }

    m.Token = document.Security.Licence
    m.UserID = m.extractUserID(document.Security.Licence)
    return nil
}

func (m *AuthMiddleware[Req, Res]) isValidToken(token string) bool {
    // Token doğrulama mantığınız
    return token != ""
}

func (m *AuthMiddleware[Req, Res]) extractUserID(token string) string {
    // Token'dan user ID çıkarma mantığınız
    return "user123"
}

// SetSelfRunables ile middleware'i aktifleştir
func (m *AuthMiddleware[Req, Res]) SetSelfRunables() error {
    m.AddRunable(m.ValidateToken)
    return nil
}

// Auth gerektiren transaction
type SecureTransaction struct {
    AuthMiddleware[SecureRequest, SecureResponse]
}

type SecureRequest struct {
    Data string `json:"data"`
}

type SecureResponse struct {
    Result string `json:"result"`
    UserID string `json:"user_id"`
}

func (t *SecureTransaction) Transact() error {
    t.Response.Result = "Processed: " + t.Request.Data
    t.Response.UserID = t.UserID // Middleware'den gelen user ID
    return nil
}
```

### Request Chaining / Zincirleme İstekler

```go
// Ana request
{
    "department": "UserService",
    "transaction": "create-user",
    "form": {
        "name": "John Doe",
        "email": "john@example.com"
    },
    "chain_request_option": {
        "user_id": "id"
    },
    "dispatchings": [
        {
            "department": "EmailService",
            "transaction": "send-welcome-email"
        },
        {
            "department": "LogService",
            "transaction": "log-user-creation"
        }
    ],
    "security": {
        "licence": "your-jwt-token"
    }
}
```

## 📚 API Dokümantasyonu / API Documentation

Framework otomatik olarak API dokümantasyonu sağlar:

- **Full Documentation**: `GET /help`
- **JSON Documentation**: `GET /help?format=json`
- **Toon Documentation (Plain Text)**: `GET /help?format=toon`
- **Short Documentation**: `GET /help?short=1`

Her transaction için dokümantasyon, teknik sözleşme ile doğrudan gönderilebilir örneği ayrı alanlarda sunar:

```json
{
    "name": "search",
    "request": {
        "schema": {
            "type": "object",
            "required": ["query"],
            "properties": {
                "query": {
                    "type": "string",
                    "minLength": 1,
                    "x-allow-empty": false
                }
            }
        },
        "example": {
            "department": "Product",
            "transaction": "search",
            "form": {
                "query": "laptop"
            }
        }
    },
    "response": {
        "schema": {
            "type": "object"
        }
    }
}
```

- `request.example`, API'ye gönderilebilen `department` / `transaction` / `form` zarfıdır.
- `request.schema` ve `response.schema`, `type`, `properties`, `required` ve validation kısıtlarını açıklayan bilgi alanlarıdır; request payload'ına eklenmez.
- Örnek değerler için struct alanlarında `example` tag'i kullanılabilir: `Query string \`json:"query" example:"laptop"\``. Tag verilmezse alana uygun temsili bir değer üretilir.

### Validation / Otomatik Doğrulama

Framework, struct tag'lerini kullanarak otomatik form doğrulaması sağlar:

```go
type UserRequest struct {
    Username string `json:"username" require:"true" is_empty:"false"`
    Email    string `json:"email" require:"true"`
    Age      int    `json:"age" require:"false"`
}
```

- `require:"true"`: JSON alanının istekte bulunmasını zorunlu kılar.
- `is_empty:"false"`: Alan gönderildiyse `null`, zero-value, boş koleksiyon/nesne veya boş/yalnız boşluk içeren string olmasını engeller. Optional alanlarda da çalışır.
- İç içe struct, pointer, slice, array ve map değerleri recursive doğrulanır; hatalar `items[0].sku` gibi tam alan yolunu içerir.
- JSON değerleri request struct'ının Go tipleriyle uyuşmuyorsa middleware ve transaction çalıştırılmaz.
- `json` tag'i bulunmayan alanlarda Go alan adı kullanılır; `json:"-"` alanları doğrulama dışında kalır.

Önceki sürümlerde kullanılan `isEmpty` yazımı geriye uyumluluk için desteklenir; yeni kodda `is_empty` kullanılmalıdır. `require` ve `is_empty` değerleri geçerli boolean (`true` veya `false`) değilse request bir yapılandırma hatasıyla reddedilir.

## 🔒 Güvenlik / Security

### Licence Validation

```go
type SecureTransaction struct {
    middleware.Middleware[Request, Response]
}

func (t *SecureTransaction) SetSelfRunables() error {
    t.AddRunable(func(document model.Document) error {
        if document.Security == nil {
            return errors.New("security required")
        }

        if !isValidLicence(document.Security.Licence) {
            return errors.New("invalid licence")
        }

        return nil
    })
    return nil
}
```

## 📝 İstemci İsteği Formatı / Client Request Format

```json
{
    "department": "DepartmentName",
    "transaction": "transaction-name",
    "form": {
        "field1": "value1",
        "field2": "value2"
    },
    "security": {
        "licence": "your-token-here"
    },
    "options": {
        "security": {
            "licence_checker": true
        }
    }
}
```

## 🔍 Response Formatı / Response Format

### Başarılı Response / Success Response
```json
{
    "department": "DepartmentName",
    "transaction": "transaction-name",
    "type": "Result",
    "output": {
        "result": "success data"
    }
}
```

### Hata Response / Error Response
```json
{
    "department": "DepartmentName",
    "transaction": "transaction-name",
    "type": "Error",
    "error": "error message"
}
```

## 📊 Logging

Framework otomatik olarak tüm request/response'ları loglar:

```json
{
    "timestamp": "2024-01-01T12:00:00Z",
    "request": {
        "method": "POST",
        "url": "/",
        "headers": {...},
        "body": {...}
    },
    "response": {
        "status_code": 200,
        "headers": {...},
        "body": {...}
    },
    "duration": "15ms"
}
```

## 🏗️ Proje Yapısı / Project Structure

```
your-project/
├── main.go
├── department/
│   ├── user/
│   │   ├── user.go
│   │   └── register.go
│   ├── product/
│   │   ├── product.go
│   │   └── register.go
│   └── auth/
│       ├── auth.go
│       └── register.go
├── middleware/
│   └── custom.go
└── model/
    └── types.go
```

## 🤝 Katkıda Bulunma / Contributing

1. Fork the project
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 Lisans / License

Bu proje MIT lisansı altında lisanslanmıştır. Detaylar için [LICENSE](LICENSE) dosyasına bakın.

## 🆘 Destek / Support

- GitHub Issues: [Create an issue](https://github.com/godispatcher/dispatcher/issues)
- Documentation: [Wiki](https://github.com/godispatcher/dispatcher/wiki)

---

**GoDispatcher** ile hızlı, güvenli ve ölçeklenebilir HTTP servisleri oluşturun! 🚀

## 🔊 Persistent Stream API (TCP/NDJSON)

HTTP bağlantı kurulum gecikmesini azaltmak için WebSocket benzeri kalıcı bir bağlantı olarak TCP tabanlı Stream API eklendi.

- Protokol: NDJSON (satır sonu ile ayrılmış JSON)
- Port: HTTP portunun +1'i (ör: HTTP 9000 ise Stream 9001)
- Her satır bir `model.Document` isteği ve tek satır JSON cevap.

Hızlı deneme (netcat):

```bash
nc localhost 9001
{"department":"Product","transaction":"getA","form":{}}
```
