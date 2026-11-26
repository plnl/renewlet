package main

// domain_lookup.go 提供域名到期日查询。
// 走后端代理 RDAP，浏览器不直连第三方，天然复用 Docker 的 HTTP(S)_PROXY 配置与超时边界。

import (
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/pocketbase/pocketbase/core"
)

type domainLookupRequest struct {
	Domain string `json:"domain"`
}

type domainLookupResponse struct {
	Domain     string `json:"domain"`
	ExpiryDate string `json:"expiryDate"`
	Registrar  string `json:"registrar,omitempty"`
	Source     string `json:"source"`
}

var domainLookupNameRe = regexp.MustCompile(`^(?i)([a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z]{2,63}$`)
var domainLookupDateRe = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}`)

const domainLookupTimeout = 12 * time.Second

func normalizeDomainLookupInput(raw string) (string, bool) {
	value := strings.ToLower(strings.TrimSpace(raw))
	value = strings.TrimPrefix(value, "http://")
	value = strings.TrimPrefix(value, "https://")
	value = strings.TrimPrefix(value, "www.")
	if index := strings.IndexAny(value, "/?#"); index >= 0 {
		value = value[:index]
	}
	value = strings.TrimSuffix(value, ".")
	if len(value) > 253 || !domainLookupNameRe.MatchString(value) {
		return "", false
	}
	return value, true
}

type rdapEvent struct {
	EventAction string `json:"eventAction"`
	EventDate   string `json:"eventDate"`
}

type rdapEntity struct {
	Roles      []string `json:"roles"`
	VcardArray []any    `json:"vcardArray"`
}

type rdapDomainPayload struct {
	ObjectClassName string      `json:"objectClassName"`
	LDHName         string      `json:"ldhName"`
	Events          []rdapEvent `json:"events"`
	Entities        []rdapEntity `json:"entities"`
}

func handleDomainLookup(_ core.App, e *core.RequestEvent) error {
	locale := requestLocale(e.Request)
	body, err := decodeStrictJSON[domainLookupRequest](e.Request, locale)
	if err != nil {
		return e.BadRequestError(validationErrorMessage(locale, "common.invalidRequestBody", err), err)
	}
	domain, ok := normalizeDomainLookupInput(body.Domain)
	if !ok {
		return e.BadRequestError("Invalid domain name", nil)
	}

	headers := http.Header{}
	headers.Set("Accept", "application/rdap+json, application/json")
	headers.Set("User-Agent", "Renewlet-DomainLookup/1.0")
	response, err := sendUpstreamRequestBytes(
		http.MethodGet,
		"https://rdap.org/domain/"+domain,
		headers,
		nil,
		upstreamHTTPRequestOptions{Provider: "rdap", Timeout: domainLookupTimeout},
	)
	if err != nil {
		return e.BadRequestError("Domain lookup failed. Please try again later.", nil)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return e.BadRequestError("Domain lookup failed. The registry did not return RDAP data.", nil)
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return e.BadRequestError("Domain lookup failed. Please try again later.", nil)
	}
	var payload rdapDomainPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return e.BadRequestError("Domain lookup failed. Unexpected RDAP response.", nil)
	}
	expiryDate := ""
	for _, event := range payload.Events {
		if strings.EqualFold(event.EventAction, "expiration") {
			if matched := domainLookupDateRe.FindString(event.EventDate); matched != "" {
				expiryDate = matched
				break
			}
		}
	}
	if expiryDate == "" {
		return e.BadRequestError("No expiration date was published for this domain.", nil)
	}
	registrar := ""
	for _, entity := range payload.Entities {
		hasRegistrarRole := false
		for _, role := range entity.Roles {
			if strings.EqualFold(role, "registrar") {
				hasRegistrarRole = true
				break
			}
		}
		if hasRegistrarRole {
			registrar = registrarNameFromVcard(entity.VcardArray)
			if registrar != "" {
				break
			}
		}
	}
	return apiSuccessJSON(e, http.StatusOK, domainLookupResponse{
		Domain:     domain,
		ExpiryDate: expiryDate,
		Registrar:  registrar,
		Source:     "RDAP",
	})
}

func registrarNameFromVcard(vcardArray []any) string {
	if len(vcardArray) < 2 {
		return ""
	}
	entries, ok := vcardArray[1].([]any)
	if !ok {
		return ""
	}
	for _, entry := range entries {
		fields, ok := entry.([]any)
		if !ok || len(fields) < 4 {
			continue
		}
		key, _ := fields[0].(string)
		if strings.EqualFold(key, "fn") {
			value, _ := fields[3].(string)
			return value
		}
	}
	return ""
}
