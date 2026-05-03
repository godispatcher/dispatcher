# Rate Limiter Altyapısı ve Tasarımı

Bu doküman, API sistemimiz için tasarlanan ve geliştirilen esnek rate limiting altyapısını açıklamaktadır.

## 1. Mimari Açıklama

Rate limiter, sistemimizi kötüye kullanıma karşı korumak, kaynak kullanımını dengelemek ve farklı kullanıcı/istemci grupları için farklı kota politikaları uygulamak amacıyla tasarlanmıştır.

### Seçilen Algoritma: Token Bucket
Bu sistemde **Token Bucket** algoritması seçilmiştir. 

**Neden Token Bucket?**
* **Burst Kapasitesi:** Kullanıcılara kısa süreli yoğun istek yapma (burst) imkanı tanır, ancak uzun vadede ortalama hızı (rate) sabit tutar.
* **Bellek Verimliliği:** Her anahtar (key) için sadece birkaç değişken (mevcut token sayısı ve son yenileme zamanı) saklanır.
* **Esneklik:** Window-based algoritmaların aksine, token'lar zamanla sürekli olarak (refill rate) eklenir, bu da daha pürüzsüz bir sınırlama sağlar.

## 2. Desteklenen Scope'lar

Rate limit aşağıdaki seviyelerde (scope) uygulanabilir:

*   **Global (`global`):** Tüm sistem genelinde uygulanan limit.
*   **IP Bazlı (`ip`):** İstek yapan istemcinin IP adresine göre uygulanan limit.
*   **User ID Bazlı (`user`):** `verify_code` üzerinden kimliği doğrulanmış kullanıcıya özel limit.
*   **API Key Bazlı (`api_key`):** `licence` (token/key) üzerinden istemci bazlı limit.
*   **Endpoint / Route Bazlı (`route`):** Belirli bir department ve transaction kombinasyonuna özel limit.
*   **Kombinasyonlu:** `GenerateKey` fonksiyonu sayesinde `user + endpoint` veya `apikey + route` gibi spesifik anahtarlar üretilebilir.

## 3. Konfigürasyon Örneği

Sistem, `TransactionOptions` üzerinden konfigüre edilebilir. YAML veya JSON formatında tanımlanabilir:

```yaml
# Örnek: Login endpoint'i için IP bazlı sıkı limit
rate_limiter:
  enabled: true
  limit: 5
  window: 60      # 60 saniyede 5 istek
  scope: "ip"

# Örnek: Search endpoint'i için API Key bazlı limit
rate_limiter:
  enabled: true
  limit: 100
  window: 3600    # Saatte 100 istek
  scope: "api_key"
```

## 4. HTTP Response ve Header'lar

Limit aşıldığında sistem otomatik olarak `429 Too Many Requests` hatası döner ve aşağıdaki standart header'ları ekler:

*   `X-RateLimit-Limit`: Tanımlanan toplam limit.
*   `X-RateLimit-Remaining`: Mevcut window içinde kalan istek hakkı.
*   `X-RateLimit-Reset`: Limitin tamamen sıfırlanacağı Unix timestamp.
*   `Retry-After`: Tekrar denemek için beklenmesi gereken saniye.

**Örnek Hata Yanıtı:**
```json
{
  "department": "Auth",
  "transaction": "login",
  "type": "Error",
  "error": "Rate limit exceeded. Try again in 12 seconds."
}
```

## 5. Ölçeklenebilirlik ve Redis Önerisi

Mevcut yapı **In-Memory** (thread-safe map) olarak çalışmaktadır ve tek instance senaryoları için uygundur. Çoklu instance (distributed) senaryolar için merkezi bir sayaç yapısı (Redis) gereklidir.

### Redis Tabanlı Yapı Önerisi:
`utilities/rate_limiter.go` içinde bir `Store` interface'i tanımlanarak Redis entegrasyonu sağlanabilir:

```go
type RateLimitStore interface {
    Take(key string, limit int, window int) (*RateLimitResult, error)
}
```

**Redis Lua Script Uygulaması:**
Atomik işlem sağlamak için Redis üzerinde şu Lua script'i kullanılabilir:
```lua
local key = KEYS[1]
local limit = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local now = tonumber(ARGV[3])

local last_tokens = tonumber(redis.call("HGET", key, "tokens") or limit)
local last_refill = tonumber(redis.call("HGET", key, "last_refill") or now)

local elapsed = math.max(0, now - last_refill)
local refill_rate = limit / window
local current_tokens = math.min(limit, last_tokens + (elapsed * refill_rate))

if current_tokens >= 1 then
    current_tokens = current_tokens - 1
    redis.call("HMSET", key, "tokens", current_tokens, "last_refill", now)
    redis.call("EXPIRE", key, window)
    return {1, current_tokens} -- Allowed
else
    return {0, 0} -- Denied
end
```

## 6. Örnek Senaryolar

1.  **IP Bazlı Login Koruması:** `scope: "ip"`, `limit: 5`, `window: 300` (5 dakikada 5 deneme).
2.  **API Key Bazlı Genel Limit:** `scope: "api_key"`, `limit: 1000`, `window: 3600`.
3.  **User + Endpoint Bazlı Özel Limit:** `scope: "user"`, `limit: 50`, `window: 60`.
4.  **Global Sistem Koruması:** `scope: "global"`, `limit: 5000`, `window: 1`.
