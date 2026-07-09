# AGENTS.md

Guidance for AI coding assistants (Claude Code, Gemini, Cursor, etc.) working in this
repository. Read this before making changes.

## What this library is

`yaaf-common-net` is a Go library providing networking utilities and a batteries-included
web server. It is the networking layer of the `go-yaaf` ecosystem and is consumed as a
dependency by other services — it is **a library, not an application** (the runnable code
under `examples/` is illustrative only).

Module path: `github.com/go-yaaf/yaaf-common-net` (Go 1.24).

Core capabilities:
- **REST server** built on `gin-gonic/gin` with a fluent builder API, group middleware,
  API-key + JWT auth, role guards, CORS, reverse proxy, and static/SPA fallback serving.
- **WebSocket server** with per-group client registries, broadcast, op-code message
  routing, and pluggable JSON decoding.
- **Utilities**: JWT token + API-key crypto, IP geolocation (IP2Location), reverse DNS,
  and well-known port lookups.

## Package layout

| Path | Purpose |
|------|---------|
| `web/` | REST + WebSocket server. Entry point: `web.NewWebServer()` in `server.go`. |
| `web/base_endpoint.go` | `BaseEndPoint` — embed it in your endpoints for param/token helpers. |
| `web/base_def.go` | `RestEntry` / `RestEndpoint` types. |
| `web/web_socket_*.go` | WebSocket listener, client, registry, message types. |
| `utils/token_utils.go` | JWT create/parse + API-key encrypt/decrypt (`TokenUtils()` singleton). |
| `utils/ip-utils.go` | IP geolocation (`IPUtils(apiKey)`), bulk lookup, reverse DNS. |
| `utils/port-utils.go` | Well-known port name/description lookup (`PortUtils()` singleton). |
| `model/` | Shared data types: `TokenData`, `IPGeoAddress`, `IPGeoPoint`. |
| `test/` | `testify`-based unit tests (package `test`). |
| `examples/` | Standalone example servers — reference only, not part of the library API. |

## Key dependencies (do not casually bump)

- `github.com/gin-gonic/gin` — HTTP framework.
- `github.com/go-yaaf/yaaf-common` — sibling library: `logger`, `entity` (e.g.
  `entity.Now()`, `entity.Timestamp`), `utils/collections`. Prefer these over
  reinventing helpers.
- `github.com/golang-jwt/jwt/v5` — **PINNED to v5.2.2**. `go.mod` explicitly warns not to
  upgrade (later versions have breaking changes). Do not change this version.
- `github.com/gorilla/websocket`, `github.com/google/uuid`,
  `github.com/ip2location/ip2location-io-go`.

## Conventions in this codebase

- **Singletons via `sync.Once`**: `TokenUtils()`, `PortUtils()` return process-wide
  singletons. `NewWebServer()` sets a package-level `serverInst` that middleware reads —
  the server is effectively a singleton too. Be aware of this shared/global state.
- **Fluent builders**: `Server` methods (`WithAppName`, `WithSecrets`, `WithAPIVersion`,
  `WithReverseProxy`, `AddRESTEndpoints`, …) return `*Server` for chaining.
- **Endpoints**: implement `web.RestEndpoint` (`Path()` + `RestEntries()`) and embed
  `web.BaseEndPoint` to inherit param parsing (`GetParamAsInt`, `GetParamAsEnum`,
  `GetParamAsTimestamp`, …) and `GetTokenData(c)`.
- **WebSocket endpoints**: implement `web.IWSEndpointConfig` (`Group()`, `Path()`,
  `WSEntries()`); route messages by integer `OpCode`; register message factories with
  `web.AddMessageFactory`.
- **Auth model**: middleware validates `X-API-KEY` (app identity) and `X-ACCESS-TOKEN`
  (JWT). `RestEntry.Skip` bit-flags bypass checks (`APIKEY=1`, `TOKEN=3`); `RestEntry.Role`
  is a bitmask matched against `TokenData.SubjectRole`. The token is re-issued with a
  fresh 30-min expiry on every authenticated response.
- Region-style `// region ... // endregion` comment blocks group related code — keep them.
- Doc comments start with the identifier name (standard Go style). Match the surrounding
  comment density and naming when editing.

## Build / test / verify

```bash
go build ./...        # compile everything
go vet ./...          # static checks
go test ./test/...    # unit tests (token/crypto tests run offline)
gofmt -l .            # must print nothing; format before committing
```

Always run `go build ./...` and `go test ./test/...` after changes. There is a GitHub
Actions build workflow (`.github/workflows/build.yml`) — keep it green.

## Security-sensitive areas — handle with care

This library is on the auth/crypto path. When touching these files, prefer secure-by-default
behavior and never reintroduce hardcoded credentials:

- **`utils/token_utils.go`** — JWT signing key and AES key are **empty by default** and
  MUST be supplied by the app via `TokenUtils().WithSecrets(apiSecret, signingKey)` (both
  ≥ 32 chars). The library **fails closed** (`ensureKeys()` panics) if they are unset. Do
  not add fallback/default secrets. JWT parsing pins the algorithm to HS256; API-key crypto
  uses authenticated **AES-GCM**. Changing the crypto format invalidates previously issued
  API keys.
- **`utils/ip-utils.go`** — the IP2Location API key must be passed by the caller
  (`IPUtils(apiKey)`); lookups return an error when it is empty. Do **not** hardcode a key.
- **`web/web_socket_listener.go`** — WebSocket `CheckOrigin` allows same-host and
  no-Origin (non-browser) clients only; additional origins must be allowlisted via
  `web.SetAllowedWSOrigins(...)`. Do not revert to "allow all origins".
- **`web/server.go`** — CORS does **not** default `Access-Control-Allow-Origin` to `*`
  (credentials/token headers are exposed). Allowed origins are reflected explicitly or set
  via `Server.WithHeader`. `customRecovery` must not echo panic details to the response.
- **`web/base_endpoint.go`** — `ResolveRemoteIp` reads `X-Forwarded-For`, which is
  client-spoofable; do not use it for security decisions without a trusted proxy chain.

If you change auth, crypto, CORS, or origin handling, call it out explicitly in your
summary so a human can review it.

## Things to avoid

- Introducing new global mutable state beyond the existing singleton pattern.
- Bumping `golang-jwt/jwt/v5` off v5.2.2.
- Adding hardcoded secrets, API keys, or credentials anywhere (including tests/examples —
  use obvious placeholders like the existing `"put your secret string..."`).
- Treating `examples/` as part of the public API surface.
