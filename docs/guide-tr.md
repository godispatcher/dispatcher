# GoDispatcher Kılavuz (TR)

Bu belge, GoDispatcher ile hızlıca başlayıp üretim seviyesinde servisler geliştirmenize yardım eder. README kısa ve odaklı tutulmuştur; detaylar bu dosyaya taşınmıştır.

## Gereksinimler
- Go 1.23+

## Hızlı Başlangıç
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
- API Dokümantasyonu: http://localhost:9000/help (kısa: /help?short=1)
- Stream (TCP/NDJSON): 9001 (HTTP portu + 1)

3) Örnek istekler (curl):

- Product.getA
```bash
curl -s http://localhost:9000/ \
  -H 'Content-Type: application/json' \
  -d '{"department":"Product","transaction":"getA","form":{}}'
```

- Product.getB
```bash
curl -s http://localhost:9000/ \
  -H 'Content-Type: application/json' \
  -d '{"department":"Product","transaction":"getB","form":{}}'
```

- Auth.login (örnek proxy mantığı ile başka servise çağrı içerir)
```bash
curl -s http://localhost:9000/ \
  -H 'Content-Type: application/json' \
  -d '{"department":"Auth","transaction":"login","form":{"username":"demo","password":"demo"}}'
```

Not: Auth.login örneği, main repo içinde dış adrese (http://localhost:1306/auth) çağrı yapan bir örnek akış içerir. Lokalinizde o servis yoksa hata almanız normaldir; Product.* istekleri lokal örnek üzerinden çalışır.

## Temel Kavramlar
- Department: İlgili transaction’ların mantıksal grubu. Örn: "Auth", "Product".
- Transaction: Tek bir iş akışı/işlem. Request/Response tipleri ile type-safe çalışır.
- Middleware: Transaction öncesi/sonrası koşan, doğrulama/zenginleştirme yapıları.

## İstek/Response Şeması
- İstek (model.Document):
```json
{
  "department": "DepartmentName",
  "transaction": "transaction-name",
  "form": {"field": "value"},
  "security": {"licence": "token-or-jwt"},
  "options": {"security": {"licence_checker": true}}
}
```

- Başarılı cevap:
```json
{
  "department": "DepartmentName",
  "transaction": "transaction-name",
  "type": "Result",
  "output": {"...": "..."}
}
```

- Hatalı cevap:
```json
{
  "department": "DepartmentName",
  "transaction": "transaction-name",
  "type": "Error",
  "error": "message"
}
```

## CORS
main.go içinde CORS örnek yapılandırması mevcuttur:
```go
service := department.NewRegisteryDispatcher("9000")
service.CORS = (&model.CORSOptions{ /* ... */ }).WithDefaults()
```
- Same-origin zorlaması gerekiyorsa `EnforceSameOrigin: true` olarak ayarlayın.

## Stream API (TCP/NDJSON)
- Amaç: HTTP bağlantı kurulum gecikmesini azaltmak için kalıcı hat.
- Port: HTTP + 1 (9000 -> 9001)
- Her satır bir JSON istek ve tek satır JSON cevap.

Deneme:
```bash
nc localhost 9001
{"department":"Product","transaction":"getA","form":{}}
```

## Kendi Department ve Transaction’ınızı Eklemek
1) Transaction tipleri:
```go
type HelloReq struct { Name string `json:"name"` }
type HelloRes struct { Message string `json:"message"` }

// Transaction
type Hello struct {
    middleware.Middleware[HelloReq, HelloRes]
}

func (t *Hello) Transact() error {
    t.Response.Message = "Hello, " + t.Request.Name + "!"
    return nil
}
```

2) Register:
```go
func Reg() {
    creator.NewTransaction[Hello, *Hello]("Greeting", "hello", nil, nil)
}
```

3) main.go’da `Reg()` çağrısı yapın.

## Güvenlik ve Lisans Doğrulama
- `document.security.licence` alanını middleware’de kontrol edebilirsiniz.
- Gelişmiş örnekler için advanced belgesine bakın.

## SSS
- Neden root path `/`? Basit gateway tarzı, tek endpoint üzerinden department/transaction yönlendirmesi.
- Dokümantasyon nasıl oluşuyor? Kayıtlı transaction’lar üzerinden tip analizi ile `/help`.

Daha fazla: docs/advanced.md
