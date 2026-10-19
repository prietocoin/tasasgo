# Diagnóstico Técnico & Plan de Migración — Tasashub → Go (Fiber)

> **Fecha:** 2026-10-17 | **Versión analizada:** Node.js / Express (actual)

---

## 1. Flujo de Despacho a WhatsApp

### Helper interno: `enviarMediaEvolution()` — `tasas.controller.js:35–75`

Todo envío a WhatsApp pasa por este helper. Construye la llamada HTTP a Evolution API:

```
POST ${EVOLUTION_API_URL}/message/sendMedia/${EVOLUTION_INSTANCE}
Headers: apikey | Global-Api-Key | Authorization: Bearer <key>
Body:    { number, mediatype, mimetype, fileName, media (base64 puro), caption }
```

---

### Endpoint `POST /api/v1/tasas/disparar` — Socio individual

```
Frontend → { socio, modo_prueba }
  │
  ├─ DB: SELECT perfiles_glaukov WHERE nombre = socio
  ├─ DB: SELECT tasas_glaukov ORDER BY id DESC LIMIT 2   (actual + anterior)
  ├─ calculator.service.js  →  calcularTasasCartelera()
  │     └─ triangulación de tasas, tendencia up/down/equal, márgenes %
  ├─ normalizarParaRenderer()  (adapta keys para Satori)
  ├─ renderer.service.js  →  generarCarteleraPNG()
  │     └─ Satori (JSX → SVG) + Resvg (SVG → PNG Buffer)
  ├─ base64Image = pngBuffer.toString('base64')
  ├─ jidDestino = modo_prueba ? JID_GRUPO_PRUEBA : socio.id_grupo
  └─ enviarMediaEvolution(jid, base64Image, caption)
```

---

### Endpoint `POST /api/v1/tasas/publicar-lote` (alias: `POST /disparar-lote`)

```
Frontend → { modo_prueba_global }
  │
  ├─ DB: SELECT tasas_glaukov LIMIT 2
  ├─ DB: SELECT * FROM perfiles_glaukov    (todos los socios activos)
  │
  └─ for each socio where activo != false   ← loop SECUENCIAL
        ├─ calcularTasasCartelera()
        ├─ generarCarteleraPNG()
        └─ enviarMediaEvolution(jid, base64, caption)
```

---

## 2. Deuda Técnica — Variables de Entorno e Instancia WhatsApp

### 2.1 Instancia whatsapp — dónde aparece

| Archivo | Línea | Código |
|---|---|---|
| `.env.example` | 11 | `EVOLUTION_INSTANCE=Jairo` |
| `tasas.controller.js` | 10 | `const EVOLUTION_INSTANCE = process.env.EVOLUTION_INSTANCE \|\| 'GLAUK'` |
| `evolution.service.js` | 13 | `const instanceName = env.EVOLUTION_INSTANCE \|\| 'GLAUK'` |
| `src/config/env.js` | — | **AUSENTE** — no está en el objeto `env` exportado |

### 2.2 Problemas identificados

1. **`env.js` incompleto.** El objeto `env` exportado sólo contiene `PORT` y variables `DB_*`. Las variables `EVOLUTION_API_URL`, `AUTHENTICATION_API_KEY`, `EVOLUTION_INSTANCE` y `JID_GRUPO_PRUEBA` son leídas directamente con `process.env` dispersos en el controller, sin centralización.

2. **Servicio desacoplado no usado.** `evolution.service.js` es **código muerto**. El controller tiene su propio helper `enviarMediaEvolution()` que es el que ejecuta en producción. Los dos bloques de envío son incompatibles en estructura de payload (`mediaMessage` vs campos planos).

3. **Loop secuencial en lote.** `publicarLoteSolo` itera socios secuencialmente con `await` dentro de `for...of`. Con N socios, el tiempo total = suma de todos los renders + envíos. Sin concurrencia ni timeout por socio.

4. **Triple header de autenticación.** El helper envía simultáneamente `apikey`, `Global-Api-Key` y `Authorization: Bearer` — redundante, indica incertidumbre sobre cuál acepta la instancia Evolution.

---

## 3. Arquitectura Propuesta — Go (Fiber)

### 3.1 Estructura de directorios

```
tasashub-go/
├── cmd/
│   └── server/
│       └── main.go                  # Entry point: carga config, init DB, monta Fiber
├── internal/
│   ├── config/
│   │   └── env.go                   # Única fuente de verdad de variables de entorno
│   ├── db/
│   │   └── postgres.go              # Pool pgx/v5
│   ├── middleware/
│   │   └── security.go              # CORS + CSP (frame-ancestors)
│   └── modules/
│       └── tasas/
│           ├── routes.go            # app.Group("/api/v1/tasas", ...)
│           ├── handler.go           # Handlers Fiber (equivale al controller)
│           ├── calculator.go        # calcularTasasCartelera()
│           ├── renderer.go          # generarCarteleraPNG() — ver §4
│           └── evolution.go         # EvolutionClient struct + SendMedia()
├── assets/
├── public/
├── .env
└── Dockerfile
```

---

### 3.2 `internal/config/env.go`

```go
package config

import (
    "os"
    "strconv"
)

type Config struct {
    Port              string
    DBHost            string
    DBPort            int
    DBUser            string
    DBPassword        string
    DBName            string
    EvolutionAPIURL   string
    EvolutionAPIKey   string
    EvolutionInstance string // default: "GLAUK"
    JIDGrupoPrueba    string
}

func Load() *Config {
    return &Config{
        Port:              getEnv("PORT", "3002"),
        DBHost:            getEnv("DB_HOST", "localhost"),
        DBPort:            getEnvInt("DB_PORT", 5432),
        DBUser:            getEnv("DB_USER", "postgres"),
        DBPassword:        getEnv("DB_PASSWORD", ""),
        DBName:            getEnv("DB_NAME", "automat"),
        EvolutionAPIURL:   getEnv("EVOLUTION_API_URL", ""),
        EvolutionAPIKey:   getEnv("AUTHENTICATION_API_KEY", ""),
        EvolutionInstance: getEnv("EVOLUTION_INSTANCE", "GLAUK"), // configurable, default GLAUK
        JIDGrupoPrueba:    getEnv("JID_GRUPO_PRUEBA", ""),
    }
}

func getEnv(key, fallback string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return fallback
}

func getEnvInt(key string, fallback int) int {
    if v := os.Getenv(key); v != "" {
        if i, err := strconv.Atoi(v); err == nil {
            return i
        }
    }
    return fallback
}
```

---

### 3.3 `internal/modules/tasas/evolution.go`

```go
package tasas

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
)

type EvolutionClient struct {
    BaseURL  string
    APIKey   string
    Instance string // inyectado desde config.Config.EvolutionInstance
}

type SendMediaRequest struct {
    Number    string `json:"number"`
    MediaType string `json:"mediatype"`  // "image"
    MimeType  string `json:"mimetype"`   // "image/png"
    FileName  string `json:"fileName"`
    Media     string `json:"media"`      // base64 puro, sin prefijo data:
    Caption   string `json:"caption"`
}

type SendMediaResponse struct {
    Key     map[string]interface{} `json:"key"`
    Message map[string]interface{} `json:"message"`
    Status  string                 `json:"status"`
}

func (c *EvolutionClient) SendMedia(ctx context.Context, req SendMediaRequest) (*SendMediaResponse, error) {
    // Modo simulación: si no hay URL configurada, loguear y salir sin error
    if c.BaseURL == "" {
        fmt.Printf("[SIMULACIÓN EVOLUTION] → %s: %s\n", req.Number, req.Caption)
        return &SendMediaResponse{Status: "simulated"}, nil
    }

    url := fmt.Sprintf("%s/message/sendMedia/%s", c.BaseURL, c.Instance)
    body, _ := json.Marshal(req)

    httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
    if err != nil {
        return nil, err
    }
    httpReq.Header.Set("Content-Type", "application/json")
    httpReq.Header.Set("apikey", c.APIKey)

    resp, err := http.DefaultClient.Do(httpReq)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result SendMediaResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }
    return &result, nil
}
```

---

### 3.4 Tabla de Endpoints — Contratos HTTP

| Método | Ruta | Body / Params | Respuesta |
|---|---|---|---|
| `GET` | `/monitor` | — | `{ success, id_tasa, tasas }` |
| `GET` | `/socios` | — | `{ success, socios[] }` |
| `GET` | `/hoo-borrador` | — | `{ success, borrador }` |
| `PUT` | `/socios/:socio` | `{ monedas, activo, modo_prueba, rol }` | `{ success, socio }` |
| `POST` | `/guardar-lote` | `{ tasas }` | `{ success, id_tasa }` |
| `POST` | `/publicar-lote` | `{ modo_prueba_global }` | `{ success, procesados, detalle[] }` |
| `POST` | `/disparar` | `{ socio, modo_prueba }` | `{ success, socio, evolutionResponse }` |
| `POST` | `/disparar-lote` | alias de `/publicar-lote` | idem |
| `GET` | `/render/:socio` | `:socio` en path | `image/png` binario |
| `POST` | `/render` | `{ socio }` | `image/png` binario |
| `DELETE` | `/lote/:id_tasa` | `:id_tasa` en path | `{ success, message }` |
| `POST` | `/notificar-remesas` | — | `{ success, notificacion }` |
| `GET` | `/health` | — | `{ status, service, timestamp }` |

---

## 4. Análisis de Alternativas — Renderizado PNG

El componente más complejo de migrar es el renderer `Satori + Resvg` (JSX → SVG → PNG).

| Opción | Descripción | Pros | Contras | Recomendación |
|---|---|---|---|---|
| **Sidecar Node** | El renderer Node permanece como microservicio interno. Go le hace `POST /render` y recibe PNG binario. | Sin cambio visual. Migración incremental. | Arquitectura híbrida, dos procesos. | ✅ **Recomendada para MVP** |
| **`chromedp`** | Go lanza headless Chrome, renderiza HTML y captura screenshot. | Render idéntico a cualquier diseño HTML/CSS. | Requiere Chrome en el contenedor (~130 MB extra). | ✅ Para diseño complejo |
| **`wkhtmltoimage`** | Go ejecuta el binario vía `os/exec`. | Simple, sin dependencias Go adicionales. | Binario externo, lento en startup, HTML limited. | ⚠️ Solo si ya disponible |
| **`go-chart` + `gg`** | Reescribir la cartelera como gráfico nativo Go. | Puro Go, sin deps externos. Muy rápido. | Requiere rediseñar visualmente desde cero. | ⚠️ Solo si diseño cambia |
| **`canvas` (Go port)** | Bibliotecas como `fogleman/gg` para dibujo 2D. | Sin proceso externo. | Limitado para layouts complejos tipo CSS. | ❌ No recomendada |

### Estrategia de migración recomendada (dos fases)

**Fase 1 — MVP Go + Sidecar Renderer:**
```
Go/Fiber (toda la lógica de negocio, DB, Evolution)
    └─ POST http://localhost:3090/render  →  Node Renderer (Satori+Resvg)  →  PNG
```

**Fase 2 — Renderer nativo con chromedp:**
```
Go/Fiber (todo, incluyendo renderer)
    └─ chromedp → headless Chrome → screenshot PNG
```

---

## 5. Mejoras Adicionales Sugeridas

| Problema actual | Solución en Go |
|---|---|
| Loop secuencial en lote | `sync.WaitGroup` + goroutines con semáforo (max 5 concurrent) |
| Triple header de auth en Evolution | Un solo header `apikey` (confirmado en Evolution API v2) |
| `evolution.service.js` código muerto | Eliminado; existe un único `EvolutionClient` |
| `env.js` incompleto | `config.Load()` centraliza **todas** las variables |
| Sin timeout en llamadas externas | `context.WithTimeout` en cada llamada HTTP |

---

*Generado como referencia técnica para la migración Tasashub → Go/Fiber.*
