# GoDispatcher Advanced Topics

This document covers advanced patterns and internal mechanisms. See docs/guide-tr.md for a Turkish-first quickstart and basics.

## Custom Middleware Patterns

- Compose multiple runables via `AddRunable`.
- Validate security/licence, enrich context, short-circuit with errors.

Example (pseudo):
```go
type AuthMW[RQ, RS any] struct {
    middleware.Middleware[RQ, RS]
}

func (m *AuthMW[RQ, RS]) Validate(doc model.Document) error {
    if doc.Security == nil || doc.Security.Licence == "" {
        return errors.New("missing licence")
    }
    return nil
}

func (m *AuthMW[RQ, RS]) SetSelfRunables() error {
    m.AddRunable(m.Validate)
    return nil
}
```

## Coordinator and Service-to-Service Calls

`coordinator.ServiceRequest` helps you call another GoDispatcher service.

```go
req := coordinator.ServiceRequest[AuthReq, AuthRes]{
    Address: "http://localhost:1306/auth",
    Document: model.Document{ Department: "Auth", Transaction: "login" },
    Request:  authReq,
}
res, err := req.CallTransaction()
```

Notes:
- Ensure CORS/headers if calling from browser.
- Prefer stable interfaces across departments.

## Request Chaining

Use `dispatchings` and `chain_request_option` to trigger follow-up transactions and pass values from previous outputs to next inputs. The framework supports this pattern through the `model.Document` fields.

## API Documentation Generator

`server.ServJsonApiDoc()` exposes `/help`. It inspects registered transactions, then renders request/response type shapes. Add your registrations before starting the server to include them in docs.

## CORS and Same-Origin

- Wraps all requests with permissive defaults; override via `model.CORSOptions`.
- `EnforceSameOrigin` can block cross-origin calls for stricter setups.

## Persistent Stream API

- Protocol: NDJSON over TCP.
- Port: HTTP + 1.
- Each line is one `model.Document` request and one JSON line response.

## Logging

Requests and responses are logged as JSON lines (log.jsonl) using github.com/godispatcher/logger. You can provide a custom writer via `RegisterDispatcher.LoggerWriter` to forward logs elsewhere.
