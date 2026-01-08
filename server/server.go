package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/godispatcher/dispatcher/constants"
	"github.com/godispatcher/dispatcher/department"
	"github.com/godispatcher/dispatcher/middleware"
	"github.com/godispatcher/dispatcher/model"
	"github.com/godispatcher/dispatcher/transaction"
	"github.com/godispatcher/dispatcher/utilities"
)

type Server[T any, TI transaction.Transaction[T]] struct {
	Options  model.ServerOption
	Runables []middleware.MiddlewareRunable
}

func (s *Server[T, TI]) AddRunable(runable middleware.MiddlewareRunable) {
	s.Runables = append(s.Runables, runable)
}

func (Server[T, TI]) GetRequest() any {
	var ta TI = new(T)
	return ta.GetRequest()
}
func (Server[T, TI]) GetResponse() any {
	var ta TI = new(T)
	return ta.GetResponse()
}

func (s Server[T, TI]) GetOptions() model.ServerOption {
	return s.Options
}

func (s Server[T, TI]) Init(document model.Document) model.Document {
	// Store verify code in a goroutine-local context for downstream calls
	if document.Security != nil && strings.TrimSpace(document.Security.VerifyCode) != "" {
		model.SetCurrentVerifyCode(document.Security.VerifyCode)
		defer model.ClearCurrentVerifyCode()
	}
	var ta TI = new(T)
	if s.Runables != nil {
		ta.SetRunables(s.Runables)
	}

	err := ta.SetSelfRunables()

	if err != nil {
		outputErrDoc := model.Document{Department: document.Department, Transaction: document.Transaction, Error: err.Error(), Type: "Error"}
		return outputErrDoc
	}

	jsonByteData, err := json.Marshal(document.Form)
	if err != nil {
		outputErrDoc := model.Document{Department: document.Department, Transaction: document.Transaction, Error: err.Error(), Type: "Error"}
		return outputErrDoc
	}

	if err != nil {
		outputErrDoc := model.Document{Department: document.Department, Transaction: document.Transaction, Error: err.Error(), Type: "Error"}
		return outputErrDoc
	}
	validator := model.DocumentFormValidater{Request: string(jsonByteData)}
	err = validator.Validate(ta.GetRequest())
	if err != nil {
		outputErrDoc := model.Document{Department: document.Department, Transaction: document.Transaction, Error: err.Error(), Type: "Error"}
		return outputErrDoc
	}
	if ta.GetRunables() != nil {
		for _, runF := range ta.GetRunables() {
			err := runF(document)
			if err != nil {
				outputErrDoc := model.Document{Department: document.Department, Transaction: document.Transaction, Error: err.Error(), Type: "Error"}
				return outputErrDoc
			}
		}
	}
	ta.SetRequest(jsonByteData)
	err = ta.Transact()
	if err != nil {
		outputErrDoc := model.Document{Department: document.Department, Transaction: document.Transaction, Error: err.Error(), Type: "Error"}
		return outputErrDoc
	}
	document.Output = ta.GetResponse()
	document.Type = "Result"

	return document
}

type TransactionListHelper struct {
	Name      string      `json:"name"`
	Procudure interface{} `json:"procedure,omitempty"`
	Output    interface{} `json:"output,omitempty"`
}
type DepartmentListHelper struct {
	Name         string                  `json:"name"`
	Transactions []TransactionListHelper `json:"transactions"`
}

type HelperList struct {
	Departments []DepartmentListHelper `json:"departments"`
}

type ApiDocServer struct {
}

func (ApiDocServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	helperList := HelperList{}
	var nestedTypeCtrl *[]string
	for _, val := range department.DispatcherHolder {
		department := DepartmentListHelper{}
		department.Name = val.Name

		for _, v := range val.Transactions {
			transaction := TransactionListHelper{}
			transaction.Name = (*v).GetName()
			if !r.URL.Query().Has("short") || r.URL.Query().Get("short") == "0" {
				nestedTypeCtrl = &[]string{}
				transaction.Procudure = utilities.Analysis((*v).GetTransaction().GetRequest(), nestedTypeCtrl)
				nestedTypeCtrl = &[]string{}
				transaction.Output = utilities.Analysis((*v).GetTransaction().GetResponse(), nestedTypeCtrl)
			}
			department.Transactions = append(department.Transactions, transaction)
		}
		helperList.Departments = append(helperList.Departments, department)
	}

	format := r.URL.Query().Get("format")
	if format == "json" {
		response, _ := json.Marshal(helperList)
		w.Header().Add(constants.HTTP_CONTENT_TYPE, constants.HTTP_CONTENT_JSON)
		fmt.Fprint(w, string(response))
		return
	}

	// HTML Output (Default)
	w.Header().Add(constants.HTTP_CONTENT_TYPE, "text/html; charset=utf-8")
	html := `
<!DOCTYPE html>
<html lang="tr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>GoDispatcher API Help</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif; line-height: 1.5; color: #333; max-width: 1000px; margin: 0 auto; padding: 20px; background-color: #f8f9fa; }
        header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 30px; border-bottom: 2px solid #dee2e6; padding-bottom: 20px; position: sticky; top: 0; background: #f8f9fa; z-index: 100; }
        h1 { margin: 0; color: #007bff; }
        .search-container { flex-grow: 1; max-width: 400px; }
        #search-input { width: 100%; padding: 10px; border: 1px solid #ced4da; border-radius: 4px; font-size: 16px; }
        .department { margin-bottom: 20px; background: #fff; border: 1px solid #dee2e6; border-radius: 8px; overflow: hidden; box-shadow: 0 2px 4px rgba(0,0,0,0.05); }
        .department-header { background: #e9ecef; padding: 10px 20px; font-weight: bold; font-size: 1.2em; border-bottom: 1px solid #dee2e6; color: #495057; }
        .transaction { border-bottom: 1px solid #f1f1f1; }
        .transaction:last-child { border-bottom: none; }
        .accordion-btn { width: 100%; text-align: left; padding: 15px 20px; background: none; border: none; outline: none; cursor: pointer; display: flex; justify-content: space-between; align-items: center; transition: background 0.2s; }
        .accordion-btn:hover { background-color: #f1f3f5; }
        .accordion-btn.active { background-color: #e7f1ff; }
        .transaction-name { font-weight: 600; font-family: monospace; color: #d63384; }
        .accordion-icon::after { content: '\002B'; font-weight: bold; }
        .active .accordion-icon::after { content: "\2212"; }
        .panel { padding: 0 20px; background-color: white; display: none; overflow: hidden; border-top: 1px solid #f1f1f1; }
        .details-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 20px; padding: 20px 0; }
        pre { background: #f8f9fa; padding: 15px; border-radius: 5px; overflow-x: auto; border: 1px solid #e9ecef; font-size: 13px; }
        .detail-title { font-weight: bold; margin-bottom: 10px; color: #6c757d; text-transform: uppercase; font-size: 0.85em; letter-spacing: 1px; }
        .no-results { display: none; text-align: center; padding: 40px; font-size: 1.2em; color: #6c757d; }
    </style>
</head>
<body>
    <header>
        <h1>GoDispatcher API Doc</h1>
        <div class="search-container">
            <input type="text" id="search-input" placeholder="Transaction veya departman ara..." onkeyup="filterContent()">
        </div>
    </header>

    <div id="content">
        {{RANGE_DEPARTMENTS}}
    </div>

    <div id="no-results" class="no-results">
        Eşleşen sonuç bulunamadı.
    </div>

    <script>
        function filterContent() {
            const input = document.getElementById('search-input');
            const filter = input.value.toUpperCase();
            const departments = document.getElementsByClassName('department');
            const content = document.getElementById('content');
            const noResults = document.getElementById('no-results');
            let anyFound = false;

            for (let i = 0; i < departments.length; i++) {
                const dept = departments[i];
                const deptName = dept.querySelector('.department-header').innerText.toUpperCase();
                const transactions = dept.getElementsByClassName('transaction');
                let deptVisible = false;

                if (deptName.indexOf(filter) > -1) {
                    deptVisible = true;
                    // If department matches, show all its transactions or just keep filtering them?
                    // Better to still show department and then check transactions
                }

                let anyTransVisible = false;
                for (let j = 0; j < transactions.length; j++) {
                    const trans = transactions[j];
                    const transName = trans.querySelector('.transaction-name').innerText.toUpperCase();
                    if (transName.indexOf(filter) > -1 || deptName.indexOf(filter) > -1) {
                        trans.style.display = "";
                        anyTransVisible = true;
                    } else {
                        trans.style.display = "none";
                    }
                }

                if (anyTransVisible) {
                    dept.style.display = "";
                    anyFound = true;
                } else {
                    dept.style.display = "none";
                }
            }

            noResults.style.display = anyFound ? "none" : "block";
        }

        const acc = document.getElementsByClassName("accordion-btn");
        for (let i = 0; i < acc.length; i++) {
            acc[i].addEventListener("click", function() {
                this.classList.toggle("active");
                const panel = this.nextElementSibling;
                if (panel.style.display === "block") {
                    panel.style.display = "none";
                } else {
                    panel.style.display = "block";
                }
            });
        }
    </script>
</body>
</html>
`
	deptHtml := ""
	for _, dept := range helperList.Departments {
		transHtml := ""
		for _, trans := range dept.Transactions {
			procJson, _ := json.MarshalIndent(trans.Procudure, "", "  ")
			outJson, _ := json.MarshalIndent(trans.Output, "", "  ")

			transHtml += fmt.Sprintf(`
            <div class="transaction">
                <button class="accordion-btn">
                    <span class="transaction-name">%s</span>
                    <span class="accordion-icon"></span>
                </button>
                <div class="panel">
                    <div class="details-grid">
                        <div>
                            <div class="detail-title">Giriş (Request) Yapısı</div>
                            <pre>%s</pre>
                        </div>
                        <div>
                            <div class="detail-title">Çıkış (Response) Yapısı</div>
                            <pre>%s</pre>
                        </div>
                    </div>
                </div>
            </div>`, trans.Name, string(procJson), string(outJson))
		}

		deptHtml += fmt.Sprintf(`
        <div class="department">
            <div class="department-header">%s</div>
            <div class="department-body">
                %s
            </div>
        </div>`, dept.Name, transHtml)
	}

	finalHtml := strings.Replace(html, "{{RANGE_DEPARTMENTS}}", deptHtml, 1)
	fmt.Fprint(w, finalHtml)
}

func ServJsonApiDoc() {
	http.Handle("/help", ApiDocServer{})
}

// ServJsonApi starts the HTTP server and applies CORS/same-origin controls if configured
func ServJsonApi(register *department.RegisterDispatcher) {
	var handler http.Handler = register
	if register != nil && register.CORS != nil {
		handler = withCORS(handler, register.CORS)
	} else {
		// apply sensible defaults (permissive CORS) to allow external control later
		defaults := (&model.CORSOptions{}).WithDefaults()
		handler = withCORS(handler, defaults)
	}
	http.Handle("/", handler)
	log.Fatal(http.ListenAndServe(":"+register.Port, nil))
}

// withCORS wraps the given handler with CORS and optional same-origin enforcement
func withCORS(next http.Handler, opts *model.CORSOptions) http.Handler {
	options := opts.WithDefaults()
	allowedMethods := strings.Join(options.AllowedMethods, ", ")
	allowedHeaders := strings.Join(options.AllowedHeaders, ", ")
	exposeHeaders := strings.Join(options.ExposeHeaders, ", ")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if options.EnforceSameOrigin && origin != "" {
			if !sameOrigin(origin, r.Host) {
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte("forbidden: same-origin policy enforced"))
				return
			}
		}

		// Preflight handling
		if r.Method == http.MethodOptions {
			applyCORSHeaders(w, origin, options, allowedMethods, allowedHeaders, exposeHeaders, r.Header.Get("Access-Control-Request-Headers"))
			// 204 No Content for preflight
			w.WriteHeader(http.StatusNoContent)
			return
		}

		applyCORSHeaders(w, origin, options, allowedMethods, allowedHeaders, exposeHeaders, r.Header.Get("Access-Control-Request-Headers"))
		next.ServeHTTP(w, r)
	})
}

func applyCORSHeaders(w http.ResponseWriter, origin string, options *model.CORSOptions, allowedMethods, allowedHeaders, exposeHeaders, reqHeaders string) {
	w.Header().Add("Vary", "Origin")
	if origin != "" {
		if originAllowed(origin, options.AllowedOrigins) {
			if options.AllowCredentials {
				// When credentials are allowed, must echo specific origin instead of '*'
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			} else {
				// Allow all or specific
				if oneStar(options.AllowedOrigins) {
					w.Header().Set("Access-Control-Allow-Origin", "*")
				} else {
					w.Header().Set("Access-Control-Allow-Origin", origin)
				}
			}
		}
		if exposeHeaders != "" {
			w.Header().Set("Access-Control-Expose-Headers", exposeHeaders)
		}
		w.Header().Set("Access-Control-Allow-Methods", allowedMethods)
		if reqHeaders != "" {
			w.Header().Set("Access-Control-Allow-Headers", reqHeaders)
		} else if allowedHeaders != "" {
			w.Header().Set("Access-Control-Allow-Headers", allowedHeaders)
		}
		if options.MaxAge > 0 {
			w.Header().Set("Access-Control-Max-Age", fmt.Sprintf("%d", options.MaxAge))
		}
	}
}

func oneStar(allowed []string) bool {
	return len(allowed) == 1 && allowed[0] == "*"
}

func originAllowed(origin string, allowed []string) bool {
	if oneStar(allowed) {
		return true
	}
	for _, a := range allowed {
		if strings.EqualFold(a, origin) {
			return true
		}
	}
	return false
}

func sameOrigin(origin, host string) bool {
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	// Compare host:port; scheme is generally irrelevant for same-origin in simple check
	return strings.EqualFold(u.Host, host)
}
