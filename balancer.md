# Balancer — практики шлюза и рыночная карта

Operational memo для хакатонного PII-прокси (`POST /process`).
Это **не** LLM load balancer: чекер бьёт sync JSON. Из LLM-шлюзов берём
форму, не KV-cache / hedging / SSE / TPM провайдера.

Спека: [`main.md`](main.md). Реализация: [`plan.md`](plan.md). Детали дня: [`hack.md`](hack.md).

---

## 1. Что не строим

| Анти-паттерн | Почему |
|---|---|
| Клон Envoy / xDS / go-control-plane data plane | go-control-plane — control plane; на день хакатона не окупается |
| Очередь (Kafka/NATS/Redis Streams) перед sync CPU API | Добавляет хвост latency; CPU не появляется из очереди. SRE: shed, не amplify |
| fasthttp / Fiber / Gin как L7 engine | Нет HTTP/2 / streaming как у `net/http`; несовместимость с `http.Handler` экосистемой |
| KV-cache / prefix-aware routing (vLLM, SGLang, Dynamo) | Нет GPU, нет KV, нет TTFT |
| Request hedging | Дублирует detect + гонка store; чекер уже ретраит |
| Python Presidio / Natasha / GLiNER / DeepPavlov на hot path | Не держат 1000 RPS × 1s на RF-типах; копируем **паттерн**, не runtime |
| Cloud DLP (Google/AWS/Azure) | Банковский контур; данные не уходят |

---

## 2. Что строим: Ambassador sidecar

```
система-потребитель → [наш модуль] → LLM
                         ↑
              mask (pre) / demask (post)
```

| Слой | Стек | Роль | Ориентир на рынке |
|---|---|---|---|
| Door | Go `net/http` | Validate, rate limit, вызов compute, 429 | LiteLLM proxy edge; Envoy/Higress door; Bifrost (Go) hygiene |
| Engine | Go `internal/compute` | Detect → mask → vault | Presidio Analyzer/Anonymizer; Higress AI Data Masking (`restore: true`) |
| Store | Redis + in-memory | `payload_id → {original, mask}` | PCI token vault; Skyflow mapping |

Один Go-модуль, два пакета (`internal/api` + `internal/compute`), вызовы функций
без HTTP. Rust снят — весь hot path на Go.

---

## 3. Практики Go API-шлюза (2026)

Источник: stdlib, Caddy, Traefik, Cloudflare Go blogs, RFC 6585/9110, Google SRE.

| Практика | Как у нас | Источник |
|---|---|---|
| Engine | `net/http`; ServeMux Go 1.22+ или chi v5 только для middleware | [pkg.go.dev/net/http](https://pkg.go.dev/net/http), [chi](https://github.com/go-chi/chi) |
| Upstream client | Клон `DefaultTransport`; `MaxIdleConnsPerHost ≥ 32`; dial ~3s; HTTP/2 | [Transport](https://pkg.go.dev/net/http#Transport), [Caddy reverse_proxy](https://caddyserver.com/docs/caddyfile/directives/reverse_proxy) |
| Server timeouts | `ReadHeaderTimeout` (Slowloris); `IdleTimeout`; handler — через `context`, не один `Client.Timeout` | [Cloudflare timeouts](https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/) |
| 429 + Retry-After | Допустимо у жюри; пишется в статистику. Capacity **~1500**, не ровно 1000 | [RFC 6585](https://www.rfc-editor.org/rfc/rfc6585.html), [RFC 9110 Retry-After](https://www.rfc-editor.org/rfc/rfc9110.html#field.retry-after) |
| Rate limit | Token bucket (`x/time/rate` или свой); size-aware (chars/tokens payload), не «1 req = 1 token» на длинных текстах | [x/time/rate](https://pkg.go.dev/golang.org/x/time/rate); LiteLLM/Kong TPM-идея |
| LB между compute | Нужен только при N>1 подах: P2C (d=2). На хакатон — 1 compute | [Mitzenmacher](https://eecs.harvard.edu/~michaelm/postscripts/tpds2001.pdf), Traefik/Caddy `p2c` |
| Graceful shutdown | `Server.Shutdown(ctx)` | stdlib |
| Observability | HTTP duration histogram (mean/p50/p95/p99); **без** тел в спанах/логах | [OTel HTTP metrics](https://opentelemetry.io/docs/specs/semconv/http/http-metrics/) |
| GOMAXPROCS | Go 1.25+ container-aware — не ставить руками в k8s | [Go 1.25 blog](https://go.dev/blog/container-aware-gomaxprocs) |

**Не делать:** `NewSingleHostReverseProxy` / deprecated `Director` (XFF spoof);
`MaxIdleConnsPerHost=2` (default → churn); retry POST без идемпотентности;
очередь перед detect.

---

## 4. Что украсть у LLM-шлюзов

| Паттерн LLM-мира | У нас | Steal? |
|---|---|---|
| Thin L7 + fat compute | Go api → Go compute (пакеты) | **yes** |
| Token/TPM admission | Размер payload (runes/4) в bucket; headroom > пика | **partial** |
| Idempotency key | `payload_id` = ключ; lookup до detect | **yes** (инверсия: completions не ретраят, `/process` обязан) |
| Mask outbound / restore inbound | Higress AI Data Masking; Presidio deanonymize | **yes** — это продукт |
| Per-tenant policy | `/systems` admin; default на чекере (нет `system_id`) | **yes**, off critical path |
| Telemetry без PII | Cloudflare metadata-only; OTel GenAI content — opt-in и опасен | **yes** |
| KV / prefix / TTFT / hedge | — | **no** |
| TPS | `est_tokens(payload)+est_tokens(result) / wall_s` | **yes** (не decode speed) |

Ближайшие продуктовые аналоги (не зависимости): Higress AI Data Masking,
Presidio, Nightfall Firewall for AI, Private AI, Cloudflare AI Gateway DLP,
BricksLLM (Go, stale).

---

## 5. Рыночная карта требований `main.md`

Формат: **ориентир → копируем → не тащим**.

| # | Требование ТЗ | Ориентир на рынке | Копируем в Go | Не тащим |
|---|---|---|---|---|
| 1 | Встраивание в цепочку consumer→LLM | Sidecar/Ambassador ([Azure](https://learn.microsoft.com/en-us/azure/architecture/patterns/sidecar)); LiteLLM pre/post_call; Envoy | Sync `POST /process`; Go door | Очередь, отдельная шина |
| 2 | Детекция ПД (качество) | Presidio hybrid; Natasha/Yargy (RU); DeepPavlov/GLiNER — batch | Regex+checksum + dict/rules на Go | BERT/GLiNER/spaCy на `/process` |
| 3 | Мало FP (Пушкин, отделение банка) | 152-ФЗ ст.3; Presidio context/deny-list | Окно ~200 + маркеры + gazetteers | Маскировать всё PER/LOC |
| 4 | Смысл для LLM после маски | PCI partial display; Presidio `mask` | Etalon-style partial (`45** ****56`, `И. И. И.`) | Redact `[TYPE]` на чекере |
| 5 | Правила per-system | Google SDP templates; LiteLLM policies | `SystemConfig` в Redis; default-policy | OPA на день 1 |
| 6 | Онбординг потребителя | SDP template + key | `POST /systems` + README ≤5 предложений | Redeploy на каждого |
| 7 | Ошибки / деградация | Circuit breaker (Azure); 503 fail-closed к LLM | Memory cache; Redis down → не 200 на новой маске; demask miss → 5xx | Forward raw PII |
| 8 | ПД не в логах/метриках; шифрование | PCI no-log PAN; Vault Transit; Redis TLS | AES-GCM `{original,mask}`; AAD=`payload_id`; логи = types/offsets | `gen_ai.prompt` в OTel |
| 9 | Расширяемость типов | Presidio `EntityRecognizer` registry | `trait Detector` + register | Monolith match |
| 10 | RF-типы (паспорт, в/у, ИНН, карта…) | FNS ММВ-7-6/435@; MVD 773/365; ISO 7812 Luhn | Нативные checksums | Cloud infoTypes |
| 11 | Регистр + даты + «серия/номер» | Presidio IGNORECASE; Natasha DatesExtractor | `(?i)`; даты **только** с маркером DOB/выдачи | Любая дата = ПД |
| 12 | Демаска, позиции | PCI vault; Presidio decrypt | Lookup `payload_id`; exact `payload==mask` | Инверсия `*` |
| 13 | Логи + Latency/RPS/TPS | OTel HTTP duration | mean/p50/p95/p99; RPS; TPS est; `mask_ok`/`demask_ok` | TTFT |
| 14 | Allowlist систем | Envoy RBAC; API keys | Redis set + default open для чекера | mTLS день 1 |
| 15 | Качество 95% | Presidio SpanEvaluator (IoU≥0.75); чекер = Levenshtein до эталона | Dual selfcheck: F1 + etalon distance | Только «нашли хоть что-то» |
| 16 | 100 000 токенов | Linear regex; chunked NER | Regex full-doc; NER chunks 4k/200 | Neural на весь текст |
| 17 | ИБ / контур банка | ГОСТ Р 57580; 152-ФЗ | Contour, encrypt store, no cloud DLP | Сертифицированный СКЗИ в zip |
| 18 | Надёжность: идемпотентность, 429 | Idempotency-Key draft; RFC 9110 | `payload_id` cache; 429+Retry-After; **не 429 demask** | Retry без store |
| 19 | Latency ≤1с, RPS 1000 | Go `regexp` (RE2-class); Hyperscan — DPI | `regexp` + aho-corasick; limiter 1500 | Hyperscan FFI, Python re |
| 20 | Плюсы: FPE/synthetic/PIN-gate/другие УЛ | NIST FF1 (не FF3); Faker ru_RU; Skyflow «не FPE CVV» | После must; PIN iff card | FF3, FPE на CVV |
| 21 | Zip / ~6к правил качества | Sonar/AlfaSonar; Clippy | `go vet`, `gofmt`, короткие модули, zip исходников | `todo!()` в релизе |

### Checker (Приложение B) — два разных score

| Шаг | Что шлёт чекер | Метрика | Следствие |
|---|---|---|---|
| **Forward** | Исходник + новый `payload_id` | Span-based Levenshtein до **эталона** | Маски в стиле эталона — must |
| **Reverse** | **Наша** маска + тот же `payload_id` | Byte-identical original | Vault + sync put до 200 — must |

---

## 6. Профиль нагрузки (жюри + ТЗ)

| Параметр | Значение | Действие |
|---|---|---|
| Средний RPS | ~330, с разгоном | Средний режим «скучный», почти 0×429 |
| Пик | 1000 | Capacity лимитера ≥1200–1500 |
| 429 | Не ошибка, в статистике | Не 429-ить demask/retry |
| Latency | mean, p50, p95, p99; ориентир ≤1с | Смотреть все перцентили |
| Pair | `mask_ok == demask_ok` | Demask без семафора detect |

---

## 7. Стек дня (заморожен)

1. **Go `internal/api`** — JSON validate → token bucket → вызов compute → 422/429/500.
2. **Go `internal/compute`** — lookup → (miss) detect → etalon partial → AES-GCM put → 200.
3. **Redis** — correspondence + optional `SystemConfig`.
4. **Не в день 0:** FPE, Faker, Hyperscan, второй LB-алгоритм, Prompt Shield.

Подробный порядок работ — [`plan.md`](plan.md).
