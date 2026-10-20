# PLAN FASE 1 — Tasashub Go/Fiber: Base del Proyecto

> **Estado:** En revisión | **Fecha:** 2026-10-17

---

## Reglas de Arquitectura (No Negociables)

| Regla | Detalle |
|---|---|
| **SRP** | Cada archivo tiene un único motivo para cambiar |
| **Longitud** | 30–60 líneas objetivo. < 30 permitido si completo. > 60 solo con Doc Comment justificando |
| **Doc Comments** | Todo paquete, struct y función pública lleva comentario de intención ("el porqué") |

---

## Ubicación del Proyecto

```
Tasashub/
├── (archivos Node.js existentes — NO se tocan)
└── tasashub-go/              ← raíz del nuevo proyecto Go
```

---

## Estructura de Archivos

```
tasashub-go/
├── go.mod                                    (go mod init tasashub)
├── cmd/
│   └── server/
│       └── main.go                           ~38 líneas
├── internal/
│   ├── config/
│   │   └── env.go                            ~55 líneas
│   ├── db/
│   │   └── postgres.go                       ~35 líneas
│   └── modules/
│       └── tasas/
│           └── evolution/
│               ├── types.go                  ~25 líneas  (solo structs)
│               └── client.go                 ~58 líneas  (HTTP client)
```

**Total: 6 archivos · ninguno supera 60 líneas.**

---

## Responsabilidad de Cada Archivo

### `cmd/server/main.go` — ~38 líneas
- Entry point: carga `config.Load()`, instancia `fiber.New()`, llama `registerRoutes()`, llama `app.Listen()`.
- Separado en `main()` + `registerRoutes()` para respetar SRP.
- Ruta `/health` expone `cfg.EvolutionInstance` → confirma visualmente la env var activa.

### `internal/config/env.go` — ~55 líneas
- Única fuente de verdad para **todas** las env vars.
- Ningún otro paquete llama `os.Getenv` directamente.
- Variables clave:

| Env Var | Default | Notas |
|---|---|---|
| `EVOLUTION_INSTANCE` | `"GLAUK"` | Configurable sin recompilar |
| `AUTHENTICATION_API_KEY` | `""` | Sin fallback — requerida en producción |
| `EVOLUTION_API_URL` | `""` | Si vacío → modo simulación en el cliente |
| `JID_GRUPO_PRUEBA` | `""` | JID del grupo de prueba WhatsApp |

### `internal/db/postgres.go` — ~35 líneas
- Función `NewPool(ctx, cfg)` que retorna `*pgxpool.Pool` o `error`.
- Hace `Ping()` al arranque → fail-fast en lugar de errores en runtime tardío.

### `internal/modules/tasas/evolution/types.go` — ~25 líneas
- Solo structs: `SendMediaRequest` y `SendMediaResponse`.
- Separado de `client.go` para que otros paquetes importen los tipos sin tirar de `net/http`.

### `internal/modules/tasas/evolution/client.go` — ~58 líneas
- Struct `Client` con campos privados: `baseURL`, `apiKey`, `instance`.
- Constructor `New(baseURL, apiKey, instance string) *Client`.
- Método `SendMedia(ctx, req) (*SendMediaResponse, error)`.
- **Un solo header de autenticación:** `apikey` (Evolution API v2).
- **Modo simulación:** si `baseURL == ""`, loguea y retorna `{Status: "simulated"}` sin tocar la red.

---

## Dependencias Go a instalar

```
github.com/gofiber/fiber/v2
github.com/jackc/pgx/v5
github.com/joho/godotenv         (para cargar .env en desarrollo)
```

---

## Plan de Verificación

### 1. Compilación
```powershell
cd tasashub-go
go mod tidy
go build ./...
# Debe compilar sin errores
```

### 2. Endpoint /health
```powershell
# Terminal 1
go run ./cmd/server

# Terminal 2
curl http://localhost:3002/health
# Esperado: { "instance": "GLAUK", "service": "TasasHub Engine ⚡", "status": "OK", ... }
```

### 3. EVOLUTION_INSTANCE configurable
```powershell
$env:EVOLUTION_INSTANCE="JAIRO"; go run ./cmd/server
curl http://localhost:3002/health
# Esperado: "instance": "JAIRO"  ← sin recompilar
```

---

## Punto de Debate Abierto

- ¿El proyecto Go va en `tasashub-go/` (subcarpeta) o en una carpeta hermana separada fuera de `Tasashub/`?
- ¿Se agrega `godotenv` para cargar `.env` automáticamente en desarrollo, o se prefieren env vars del sistema?

---

*Este documento reemplaza cualquier plan verbal discutido en el chat. Una vez aprobado, se procede a escribir el código fuente.*
