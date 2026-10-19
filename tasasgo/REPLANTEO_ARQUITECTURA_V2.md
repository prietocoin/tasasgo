# REPLANTEO DE ARQUITECTURA V2 — `tasashub-go`

> **Estado:** Propuesta para aprobación | **Fecha:** 2026-10-17
> **Alcance:** sustituye únicamente el *reparto de archivos* de `PLAN_FASE1.md`. Las **reglas** (SRP, 30–60 líneas, Doc Comments) y el alcance funcional de la Fase 1 se mantienen.
> **Contexto de entrada:** `PLAN_FASE1.md`, `DIAGNOSTICO_MIGRACION.md` y el código Node verificado en disco.
> **Restricciones de esta tarea:** sin preguntas interactivas, sin escribir código fuente. Este documento es el único artefacto.

---

## 0. Resumen ejecutivo

1. `PLAN_FASE1.md` plantea **6 archivos**: `main.go`, `config/env.go`, `db/postgres.go`, `evolution/types.go` y `evolution/client.go` (+ `go.mod`).
2. Ese reparto **no sobrevive** a la regla de 30–60 líneas una vez se añaden Doc Comments obligatorios, manejo de errores con envoltura (`%w`), timeouts HTTP y validación de status. La evidencia numérica está en §3.
3. Esta V2 propone **13 archivos** para la misma Fase 1 (mismo comportamiento, mismo alcance), separados por *motivo de cambio* y no por cantidad de líneas.
4. La proyección del proyecto completo (las 12 rutas, el calculador, el renderer y el store) da **~45 archivos Go**, y ese número es la consecuencia directa —y deseable— del límite de 60 líneas.
5. Se detectaron **6 hallazgos bloqueantes** que no son de reparto (Go no instalado, clave de API expuesta, ausencia de control de versiones, conflicto de puerto, `client.go` que excede el límite, discrepancia entre documentos). Están en §11.

---

## 1. Inventario real verificado del proyecto actual

Medición obtenida en disco con `Get-Content | Measure-Object -Line` (cuenta **líneas con contenido**, excluye líneas en blanco).

| Archivo | Líneas con contenido | Rol actual |
|---|---:|---|
| `server.js` | 29 | Bootstrap Express, CORS, CSP, estáticos, `/health` |
| `src/config/env.js` | 10 | Solo `PORT` y `DB_*` |
| `src/config/db.js` | 17 | Pool `pg` |
| `src/modules/tasas/routes/tasas.routes.js` | 33 | Tabla de 13 rutas |
| `src/modules/tasas/controllers/tasas.controller.js` | 379 | 12 handlers + 2 helpers + 3 constantes de entorno |
| `src/modules/tasas/services/calculator.service.js` | 86 | Triangulación, márgenes, tendencia, fecha/hora |
| `src/modules/tasas/services/db.service.js` | 61 | 4 consultas — **aparentemente sin uso** (el controller consulta `pool` en línea) |
| `src/modules/tasas/services/evolution.service.js` | 35 | **Código muerto** con payload incompatible |
| `src/modules/tasas/services/renderer.service.js` | 87 | Satori + Resvg + caché de assets |
| `src/modules/tasas/templates/cartelera.template.js` | 315 | JSX de la cartelera |
| `src/utils/formatters.js` | 27 | `aplicarReglaPrecisionTasa`, `normalizarNumero`, `truncarTasaOficial` |
| **Total lógica de servidor** | **1.079** | — |

Otros archivos relevantes (no son lógica de servidor):

| Archivo | Líneas | Nota |
|---|---:|---|
| `public/index.html` | 679 | Dashboard SPA servido como estático |
| `public/sw.js` | 9 | Service worker |
| `Dockerfile` | 20 | Node 20-alpine, `EXPOSE 3002` |
| `.env.example` | 12 | Contiene una clave con pinta de real (32 hex) — ver §11.2 |
| `assets/` | — | `fonts/Inter-Bold.ttf`, `flags/`, `background.jpg` (2.0 MB), `logo.png` (190 KB) |

**Consecuencia arquitectónica:** 1.079 líneas de lógica, con el techo de 60 líneas por archivo, implican un piso de ~25 archivos para el proyecto completo; con SRP realista, **~45**. La V2 no es "más archivos por gusto": es la aritmética del límite.

---

## 2. Reglas de arquitectura y su verificación mecánica

| Regla | Definición operativa en V2 |
|---|---|
| **SRP** | Un archivo = **un motivo de cambio**. Si dos funciones cambiarían por razones distintas (p. ej. "cambia el contrato de WhatsApp" vs. "cambia el algoritmo de redondeo"), van en archivos distintos aunque quepan juntas. |
| **30–60 líneas** | Se cuentan **todas** las líneas del archivo: Doc Comments, `package`, `import`, líneas en blanco y cierres. `< 30` permitido solo si el archivo está completo (sin lógica pendiente). `> 60` prohibido salvo excepción documentada con Doc Comment que justifique. |
| **Doc Comments** | Obligatorio en `package`, todo `type` y toda función exportada. Describe **la intención (el porqué)**, no la mecánica (el qué). Funciones no exportadas: comentario cuando su intención no sea evidente por el nombre. |

**Verificación (no se confía en la estimación, se mide):**

```powershell
gofmt -l .                 # debe imprimir vacío
go vet ./...               # debe salir limpio
go build ./...             # debe compilar
.\scripts\check-file-length.ps1   # debe reportar todos los .go en 30..60
```

**Política de exenciones propuesta** (requiere aprobación, §10-D1):

| Categoría | ¿Aplica 30–60? | Límite propio |
|---|---|---|
| `*.go` de producción | Sí | 30–60 |
| `*_test.go` | No | ≤ 100 |
| `go.mod`, `.env.example`, `Dockerfile`, `docker-compose.yml` | No | — (no son código) |
| `scripts/*` (herramientas) | No | ≤ 60 por higiene |

---

## 3. Auditoría del reparto de `PLAN_FASE1.md` (por qué 6 archivos no alcanzan)

### 3.1 `internal/modules/tasas/evolution/client.go` — **excede el límite**

Recuento del archivo propuesto en `DIAGNOSTICO_MIGRACION.md` §3.3, sin mejoras: **63 líneas**. Al aplicar lo que el mismo diagnóstico exige en su §5 (`context.WithTimeout` en llamadas externas, envoltura de errores, y —añadido por esta revisión— verificación de `resp.StatusCode` antes del decode) el archivo llega a **~70 líneas**. Es la única violación dura identificada.

### 3.2 `internal/config/env.go` — **cabe, pero sin margen**

Recuento con Doc Comments incluidos: **~58 líneas** (10 campos de `Config` + `Load()` + `getEnv` + `getEnvInt`). Queda dentro del techo por **dos líneas**. Cualquier variable futura ya identificada (`DB_SSLMODE`, `RENDERER_URL`, `HOO_GLOBAL_RATES_URL`, `APP_ENV`) obligaría a partirlo *en el momento menos oportuno*. La V2 lo parte **ahora**, con la superficie completa prevista.

### 3.3 `cmd/server/main.go` — **rompe SRP**

El plan asigna a `main.go` tres razones de cambio: (1) cómo se ensambla el arranque, (2) qué rutas existen, (3) la forma del payload `/health`. Sumado a que `main()` también debe crear el pool, el archivo queda en ~51 líneas: cabe, pero viola SRP. En Go esto no tiene costo de diseño: varios archivos en el mismo directorio forman **un solo paquete `main`**.

### 3.4 `internal/db/postgres.go` — **código muerto en Fase 1**

`PLAN_FASE1.md` lista `config.Load()`, `fiber.New()`, `registerRoutes()`, `app.Listen()`. **No** menciona `db.NewPool()`. Si `main` no lo llama, el pool no se crea, `/health` no prueba la base y la Fase 1 no tiene *fail-fast*. La V2 lo cablea.

### 3.5 Discrepancia entre `PLAN_FASE1.md` y `DIAGNOSTICO_MIGRACION.md`

El diagnóstico §3.1 propone `internal/modules/tasas/{routes,handler,calculator,renderer,evolution}.go`. Un `handler.go` único para **12 endpoints** (con 4 de ellos reutilizando el mismo pipeline) no cabe en 60 líneas de ninguna forma. Ese listado pertenece a fases posteriores y debe subdividirse por recurso (§7).

### 3.6 Lo que se conserva del plan original

- Estructura `cmd/` + `internal/` (layout estándar de Go, correcto).
- `tasashub-go/` como subcarpeta del proyecto actual.
- `config` como **única** fuente de variables de entorno (ningún otro paquete llama `os.Getenv`).
- `db.NewPool(ctx, cfg)` con `Ping()` al arranque.
- Separación de `evolution/types.go` de la lógica HTTP, para que otros paquetes importen el contrato sin arrastrar `net/http`. Esta decisión del plan es **correcta y poco frecuente**: es la que habilita los tests de handlers sin red.
- Un **solo** header de autenticación: `apikey`.
- **Modo simulación** como ciudadano de primera clase (permite correr el servicio completo sin WhatsApp).

---

## 4. Nueva estructura de archivos — Fase 1 (13 archivos)

```
tasashub-go/
├── go.mod                                       manifiesto
├── .env.example                                 plantilla de entorno (sin secretos)
├── cmd/
│   └── server/
│       ├── main.go                              arranque y wiring
│       ├── app.go                               construcción del servidor Fiber
│       ├── routes.go                            tabla de rutas
│       └── health.go                            contrato del endpoint /health
└── internal/
    ├── config/
    │   ├── config.go                            qué ajustes existen + Load()
    │   └── lookup.go                            primitivas de lectura de entorno
    ├── db/
    │   └── postgres.go                          ciclo de vida del pool
    └── modules/
        └── tasas/
            └── evolution/
                ├── types.go                     contrato de wire con Evolution API
                ├── client.go                    estado y construcción del cliente
                ├── send.go                      operación POST /message/sendMedia
                └── send_test.go                 pruebas del cliente (exento del techo)
```

**13 archivos** (12 de código + manifiesto). Ninguno supera 60 líneas; ninguno baja de 30 salvo `client.go`, que está completo por diseño.

### 4.1 Responsabilidad, motivo de cambio y presupuesto

| # | Archivo | Único motivo para cambiar | Líneas objetivo |
|---:|---|---|---:|
| 1 | `go.mod` | Dependencias y versión del módulo | ~10 |
| 2 | `.env.example` | Catálogo de variables documentadas | ~14 |
| 3 | `cmd/server/main.go` | Cómo se ensambla y arranca el proceso | ~30 |
| 4 | `cmd/server/app.go` | Cómo se construye el servidor HTTP (middleware, límites) | ~38 |
| 5 | `cmd/server/routes.go` | Qué rutas existen | ~26 |
| 6 | `cmd/server/health.go` | La forma (contrato) del payload `/health` | ~40 |
| 7 | `internal/config/config.go` | Qué ajustes existen y sus valores por defecto | ~42 |
| 8 | `internal/config/lookup.go` | Cómo se lee y convierte una variable de entorno | ~34 |
| 9 | `internal/db/postgres.go` | Cómo se abre y verifica el pool | ~30 |
| 10 | `evolution/types.go` | El contrato de datos con Evolution API | ~24 |
| 11 | `evolution/client.go` | Cómo se construye y qué estado guarda el cliente | ~26 |
| 12 | `evolution/send.go` | La llamada HTTP concreta `sendMedia` | ~34 |
| 13 | `evolution/send_test.go` | Comportamiento verificable del cliente | ~55 (techo 100) |

### 4.2 Contenido exacto de cada archivo

**`cmd/server/main.go` (~30)** — `main()` únicamente: `godotenv` (dev), `config.Load()`, `context` de arranque, `db.NewPool`, `log.Fatalf` si la base no responde, `newApp(cfg, pool)` (viene de `app.go`), `log.Printf` del puerto y `app.Listen`. **No** define rutas ni payloads.

**`cmd/server/app.go` (~38)** — `newApp(cfg *config.Config, pool *pgxpool.Pool) *fiber.App`: instancia `fiber.New()` con `BodyLimit` (el controller Node acepta `10mb`; hay que replicarlo para no romper cargas de imágenes), llama a `registerRoutes(app, cfg, pool)` y devuelve el `*fiber.App`. Es el único lugar donde se toca la configuración del framework.

**`cmd/server/routes.go` (~26)** — `registerRoutes(app *fiber.App, cfg *config.Config, pool *pgxpool.Pool)`: únicamente la tabla de rutas. En Fase 1 son `GET /health` (desde `health.go`) y, cuando se apruebe la decisión §10-D7, la ruta de sondeo de Evolution. Sin lógica.

**`cmd/server/health.go` (~40)** — `healthHandler(cfg *config.Config, pool *pgxpool.Pool) fiber.Handler`: devuelve la **función** handler (patrón de inyección por cierre, sin variables globales). Reporta `status`, `service`, `instance` (leído de `cfg.EvolutionInstance`) y `db` (resultado de `pool.Ping(ctx)` con timeout corto). Motivo de cambio: la forma del payload.

**`internal/config/config.go` (~42)** — `type Config struct` con todos los campos tipados y `func Load() *Config` que los resuelve usando las primitivas de `lookup.go`. Superficie prevista: `Port`, `AppEnv`, `DBHost`, `DBPort`, `DBUser`, `DBPassword`, `DBName`, `DBSSLMode`, `EvolutionAPIURL`, `EvolutionAPIKey`, `EvolutionInstance`, `JIDGrupoPrueba`, `RendererURL`. Es la **única** declaración de "qué ajustes existen".

**`internal/config/lookup.go` (~34)** — `getEnv(key, fallback string) string`, `getEnvInt(key string, fallback int) int` y (nuevo, para `DBSSLMode`/`AppEnv` si se opta por tiparlos) `getEnvBool`. Sin conocimiento del dominio: solo primitivas. Motivo de cambio: la mecánica de lectura de entorno.

**`internal/db/postgres.go` (~30)** — `NewPool(ctx, cfg *config.Config) (*pgxpool.Pool, error)`: construye el DSN, `pgxpool.New`, `Ping` y —en caso de fallo— `Close()` antes de devolver el error envuelto con `%w`. Ninguna consulta SQL: eso pertenece al futuro `store/`.

**`evolution/types.go` (~24)** — Solo `SendMediaRequest` y `SendMediaResponse`. Sin imports de `net/http`. Documenta explícitamente que `Media` es **base64 puro, sin el prefijo `data:`** (el código muerto `evolution.service.js` sí incluía el prefijo: es la causa de que ambos payloads fueran incompatibles).

**`evolution/client.go` (~26)** — Campos privados (`baseURL`, `apiKey`, `instance`, `httpClient`), constructor `New(baseURL, apiKey, instance string) *Client` con `http.Client{Timeout: 30s}` y el predicado `Simulated()`. Ninguna operación HTTP.

**`evolution/send.go` (~34)** — `SendMedia(ctx, req) (*SendMediaResponse, error)`: cortocircuito en modo simulación, marshal, URL, request con contexto, header `apikey` **único**, verificación de `resp.StatusCode >= 300` antes del decode, y envoltura de todos los errores. Motivo de cambio: el protocolo de Evolution.

**`evolution/send_test.go` (~55)** — Con `httptest.Server`: (1) `baseURL == ""` → `Status: "simulated"` y **cero** peticiones recibidas (modo simulación verificado sin red); (2) `baseURL != ""` → exactamente 1 petición, header `apikey` presente y **único** (verifica el hallazgo §2.4 del diagnóstico por la vía ejecutable); (3) instancia vacía → URL malformada detectada antes del envío. Motivo de cambio: el comportamiento esperado del cliente.

---

## 5. Cómo se conectan los archivos (grafo de dependencias)

```
                        ┌──────────────────┐
                        │ cmd/server/main  │
                        └────────┬─────────┘
              ┌──────────────────┼───────────────────┐
              ▼                  ▼                   ▼
     ┌────────────────┐  ┌──────────────┐  ┌──────────────────┐
     │ internal/config│  │  internal/db │  │ cmd/server/app   │
     │  config.go     │  │ postgres.go  │  └────────┬─────────┘
     │  lookup.go     │  └──────┬───────┘           ▼
     └────────────────┘         │          ┌──────────────────┐
              │                 │          │ cmd/server/routes│
              │                 │          └────────┬─────────┘
              │                 │                   ▼
              │                 │          ┌──────────────────┐
              │                 │          │ cmd/server/health│
              └─────────────────┴──────────┴──────────────────┘
                                     (solo tipos)

     ┌────────────────────────────────────────────────────────┐
     │ evolution/  types.go ◄── client.go ◄── send.go         │   ← NO cableado en Fase 1
     │             (solo stdlib; jamás importa paquetes tasas) │   ← se ejercita por send_test.go
     └────────────────────────────────────────────────────────┘
```

**Reglas de acoplamiento que se respetan en toda la migración:**

1. **Dirección única:** `cmd/` → `internal/...`. Ningún paquete `internal/` importa a `cmd/` (imposible por diseño en Go) y `evolution` **nunca** importa el paquete `tasas` (evita ciclos cuando los handlers lo consuman).
2. **Inyección explícita, sin globales:** `cfg` y `pool` se pasan como parámetros. Cero `init()`, cero variables de paquete mutables. (El Node actual usa un `let fontBuffer` y un `const assetsCache` a nivel de módulo; en Go esas cachés vivirán **dentro** de los structs que las poseen.)
3. **`context.Context` como primer parámetro** de toda función que toque red, disco o base de datos, con timeout propio en el borde.
4. **Aceptar interfaces, devolver structs:** `evolution.Client` y el futuro `renderer.Renderer` son structs; las interfaces se declaran en el paquete **consumidor**, no en el que implementa.
5. **Errores envueltos con `%w`** y mensajes en formato `verbo_objeto: causa`. Sin `panic` fuera de `main`.
6. **Contratos de datos en el paquete del contrato:** `types.go` es el único lugar donde se describe el JSON de Evolution. Si mañana cambia la API, se cambia un archivo.

---

## 6. Cómo se distribuye la lógica para no pasar de 60 líneas

La estrategia no es "cortar por la mitad", es **cortar por eje de cambio**. Cinco reglas operativas:

| Regla de corte | Ejemplo en la migración |
|---|---|
| **1 archivo por concepto de wire** | `evolution/types.go` (contrato) ≠ `send.go` (protocolo) ≠ `client.go` (estado). |
| **1 archivo por recurso HTTP, no por capa** | `handler_lote.go` (guardar/publicar/eliminar) ≠ `handler_disparo.go` (enviar a WhatsApp) ≠ `handler_lectura.go` (monitor/socios/borrador). |
| **1 archivo por regla de negocio** | `calculator/precision.go` (regla de redondeo a 1e-8) ≠ `calculator/triangulacion.go` (márgenes) ≠ `calculator/tendencia.go` (up/down/equal) ≠ `calculator/tiempo.go` (zona horaria). |
| **1 archivo por tabla** | `store/lotes.go` (`tasas_glaukov`) ≠ `store/socios.go` (`perfiles_glaukov`) ≠ `store/notificaciones.go` (`notificaciones_tasas`). |
| **1 archivo por dependencia externa** | `renderer/sidecar.go` (Node) ≠ `renderer/chromedp.go` (Chrome, Fase 2). El resto del código solo conoce la interfaz. |

**Tablas de datos como `map[string]string`, no como código:** los mapas `BANDERAS_MAP` (18 entradas) y `NOMBRES_PAIS` (18 entradas) de `calculator.service.js` son ~14 líneas de código, pero son **datos**. En V2 van en `calculator/catalogo.go` como datos con Doc Comment. Si el catálogo crece, se mueve a `catalogo.go` + un archivo por familia de monedas, sin tocar el algoritmo.

**Extracción de literales:** `HOO_GLOBAL_RATES_URL`, los JIDs por defecto y los textos de caption (`🧪 [PRUEBA]`, `🚀 [OFICIAL]`) son constantes de configuración o de presentación. Salen del medio de un handler y van a `config.go` (URL) o a un `mensajes.go` de presentación. El `char` del multiplicador de precisión (`1e8`) es una constante nombrada en `precision.go`.

---

## 7. Mapa completo de archivos previstos (Fases 1–7)

Proyección de los **~45 archivos** que tendrá el proyecto. Se documenta aquí para que el reparto de la Fase 1 no deba rehacerse después.

```
tasashub-go/
├── go.mod
├── .env.example
├── Dockerfile                                    (F7) multi-stage + tzdata embebido
├── docker-compose.yml                            (F7) go + renderer + postgres
├── scripts/
│   └── check-file-length.ps1                     (F1) auditoría de 30–60 líneas
├── cmd/
│   ├── server/                                   (F1) main · app · routes · health
│   └── evoprobe/main.go                          (F1) CLI de sondeo real de Evolution (§10-D7)
└── internal/
    ├── config/                                   (F1) config.go · lookup.go
    ├── db/postgres.go                            (F1)
    ├── middleware/
    │   └── security.go                           (F5) CORS + CSP frame-ancestors (paridad con server.js)
    ├── statico/
    │   └── montaje.go                            (F5) public/ + /flags (paridad con express.static)
    └── modules/tasas/
        ├── routes.go                             (F5) grupo /api/v1/tasas
        ├── dominio/
        │   └── tipos.go                          (F2) Perfil · Lote · TarjetaPais · Cartelera
        ├── calculator/
        │   ├── precision.go                      (F2) aplicarReglaPrecisionTasa
        │   ├── number.go                         (F2) normalizarNumero
        │   ├── catalogo.go                       (F2) banderas + nombres de país
        │   ├── base.go                           (F2) tasa de la moneda nativa del socio
        │   ├── triangulacion.go                  (F2) cross-rate + márgenes D/P
        │   ├── tendencia.go                      (F2) up / down / equal
        │   ├── tiempo.go                         (F2) fecha/hora America/Caracas
        │   └── cartelera.go                      (F2) orquestador calcularTasasCartelera
        ├── store/
        │   ├── escaneo.go                        (F3) scan + unmarshal de JSONB
        │   ├── lotes.go                          (F3) tasas_glaukov
        │   ├── socios.go                         (F3) perfiles_glaukov
        │   └── notificaciones.go                 (F3) notificaciones_tasas
        ├── renderer/
        │   ├── contrato.go                       (F4) interface Renderer
        │   ├── normalizar.go                     (F4) normalizarParaRenderer
        │   ├── sidecar.go                        (F4) cliente HTTP del renderer Node
        │   └── sidecar_test.go                   (F4)
        ├── hoo/
        │   ├── cliente.go                        (F6) GET global-rates
        │   └── mapeo.go                          (F6) mapeo de ~20 campos
        ├── handler_lectura.go                    (F5) /monitor · /socios · /hoo-borrador
        ├── handler_socio.go                      (F5) PUT /socios/:socio
        ├── handler_lote.go                       (F5) /guardar-lote · /publicar-lote · DELETE /lote/:id
        ├── handler_disparo.go                    (F5) /disparar · /disparar-lote
        ├── handler_render.go                     (F5) GET /render/:socio · POST /render
        ├── handler_remesas.go                    (F5) /notificar-remesas
        ├── servicio_publicar.go                  (F5) pipeline lote → socios → envío
        ├── concurrencia.go                       (F5) worker pool (máx. 5) + timeout por socio
        └── evolution/                            (F1) types · client · send · send_test
```

**Fases y criterio de salida**

| Fase | Contenido | Criterio de salida |
|---|---|---|
| **F1** | Esqueleto: config, db, cliente Evolution, `/health`, tests del cliente | `go build ./...` limpio · `/health` responde `instance` + `db` · modo simulación probado por test |
| **F2** | Calculador puro (sin HTTP ni DB) | **Tests de caracterización**: mismos números que `calculator.service.js` sobre una matriz de fixtures |
| **F3** | `store/` — consultas tipadas | Los 4 endpoints de lectura devuelven la misma forma que hoy |
| **F4** | `renderer/` detrás de interfaz + sidecar Node | `POST /render` en Go devuelve un PNG **byte-idéntico** al actual (comparación por hash) |
| **F5** | Handlers de los 12 endpoints + concurrencia + estáticos | Paridad de contratos HTTP con `tasas.routes.js` verificada endpoint por endpoint |
| **F6** | `hoo/` + middleware de seguridad | `frame-ancestors` idéntico a `server.js` y borrador HOO con los 20 campos |
| **F7** | Docker multi-stage + compose + tzdata | Imagen < 20 MB, sin Chrome, con `America/Caracas` correcto |

---

## 8. Riesgos de migración detectados (no de reparto)

| # | Riesgo | Evidencia | Mitigación propuesta |
|---:|---|---|---|
| R1 | **El calculador es la pieza de mayor riesgo** (define tasas de dinero). El proyecto Node **no tiene un solo test**. | `package.json` no declara tests; no existe `*.test.js` | Tests de caracterización en F2: ejecutar el Node actual sobre una matriz de fixtures (socios × lotes × márgenes) y **congelar la salida JSON** como expected. El puerto en Go debe reproducirla exactamente. |
| R2 | **Formato de fecha por locale no portable.** `toLocaleDateString('es-VE', {month:'short'})` produce abreviaturas que Go no genera igual (`oct`/`oct.`, mayúsculas, punto final). | `calculator.service.js:86–97` | R1 lo cubre: las cadenas exactas quedan fijadas en los fixtures. |
| R3 | **Zona horaria en imagen mínima.** `America/Caracas` requiere tzdata; una imagen distroless/scratch sin tzdata cae a UTC y desplaza fecha/hora 4 h. | `renderer.service.js` y `calculator.service.js` usan la zona explícitamente | `import _ "time/tzdata"` en el binario final (F7) y verificarlo en el test de F2. |
| R4 | **Renderer no migrable sin pérdida visual.** `cartelera.template.js` son 315 líneas de JSX + Satori + Resvg compilado en Rust. | análisis §4 del diagnóstico | Sidecar Node tras interfaz (F4) y comparación por hash del PNG. Solo después, `chromedp` (F4b). |
| R5 | **Doble fuente de assets.** `assets/` (fuente Inter, 18 banderas, fondo 2 MB) lo seguirá usando el renderer Node. | `assets/` verificado en disco | Una sola copia compartida hasta que el renderer pase a Go; entonces se evalúa mover a `tasashub-go/assets/` con un único origen. |
| R6 | **El dashboard es un estático de 679 líneas** que llama a `/api/v1/tasas/*` y a `/flags`. | `public/index.html`, `server.js:23–24` | F5 debe replicar `public/` y el mapeo `/flags` → `assets/flags`, o mantener el estático servido por Node y solo migrar la API. Decisión §10-D13. |
| R7 | **Sin red de seguridad de contratos.** No hay tests de integración ni esquemas; los contratos solo viven en el frontend. | `tasas.routes.js` + `index.html` acoplados por convención | Congelar los 13 contratos en una tabla (ya existe en el diagnóstico §3.4) y usarla como checklist de F5. |

---

## 9. Orden de ejecución de la Fase 1

Cada paso se cierra con su verificación antes de pasar al siguiente.

| Paso | Acción | Verificación |
|---:|---|---|
| 0 | Prerequisitos: instalar Go, rotar la clave expuesta, respaldar/versionar antes de tocar nada | `go version` responde · clave nueva en gestor de secretos |
| 1 | `go mod init tasashub` + dependencias + `scripts/check-file-length.ps1` | `go mod tidy` sin errores · script audita la carpeta |
| 2 | `internal/config/lookup.go` → `internal/config/config.go` | `go vet ./internal/config/...` |
| 3 | `internal/db/postgres.go` | Compila; se ejercita en el paso 7 |
| 4 | `evolution/types.go` → `client.go` → `send.go` | `go vet ./internal/modules/...` |
| 5 | `evolution/send_test.go` | `go test ./internal/modules/tasas/evolution/...` en verde |
| 6 | `cmd/server/health.go` → `routes.go` → `app.go` → `main.go` | `go build ./...` |
| 7 | Arranque real | `/health` → `{"instance":"GLAUK","db":"ok",...}` |
| 8 | Verificación de configurabilidad | `$env:EVOLUTION_INSTANCE="JAIRO"` → `/health` refleja `JAIRO` **sin recompilar** |
| 9 | Auditoría final | `gofmt -l .` vacío · `go vet ./...` limpio · tabla de líneas real ≤ 60 · `.env.example` sin secretos |

---

## 10. Decisiones técnicas que necesito que tomes

Cada una trae opciones y **mi recomendación**. Ninguna bloquea la aprobación del resto.

| # | Decisión | Opciones | Recomendación |
|---:|---|---|---|
| **D1** | Alcance del límite 30–60 | (a) solo `*.go` de producción, tests ≤100, tooling exento · (b) todo, incluidos tests · (c) todo, incluidos `go.mod`/`Dockerfile` | **(a)**. Con (b) o (c) el límite se vuelve un objetivo en sí mismo y fuerza fragmentar tablas de test ilegibles. |
| **D2** | Versión mayor de Fiber | (a) `fiber/v2` (lo dice el plan) · (b) `fiber/v3` | **(b), previa confirmación de la versión estable vigente**. Es greenfield: adoptar el major actual ahora es barato; hacerlo en F5 no lo es. Si prefieres no desviarte del plan, (a) y se anota la deuda. |
| **D3** | Puerto por defecto | (a) `3002` (igual que Node) · (b) `3003` | **(b)**, configurable por `PORT`. Evita que Go y Node colisionen durante la coexistencia de F1–F6. |
| **D4** | Carga de `.env` (`godotenv`) | (a) siempre, ignorando error · (b) solo si `APP_ENV != production` · (c) no usarlo; solo env del sistema | **(b)**: evita que un `.env` olvidado en producción "tape" una variable mal definida, y mantiene la comodidad en local. |
| **D5** | Cablear Postgres en F1 | (a) sí, con fallo al arrancar · (b) no (literal al plan) | **(a)**. Convierte `postgres.go` en código ejercitado y da fail-fast, como pide el propio diagnóstico §5. |
| **D6** | Cómo validar el header `apikey` contra Evolution real en F1 | (a) ruta HTTP de depuración tras flag · (b) CLI de un disparo `cmd/evoprobe` · (c) solo tests con `httptest` | **(b)**: valida el hallazgo más incierto del diagnóstico **sin** ampliar la superficie HTTP pública. La opción (a) deja una ruta que hay que recordar cerrar. |
| **D7** | Default de `JID_GRUPO_PRUEBA` | (a) `""` (lo dice el plan) · (b) el grupo real, como el Node | **(a)**, y fallar con error explícito si se usa en modo prueba sin configurarlo. Un JID de producción no debe estar escrito en el código. |
| **D8** | Fallback `AUTHENTICATION_API_KEY` → `EVOLUTION_API_KEY` | (a) mantener ambos · (b) solo el primero | **(b)** salvo que el despliegue actual use el segundo nombre, en cuyo caso se documenta y se mantiene temporalmente. |
| **D9** | `DB_SSLMODE` | (a) fijo `disable` · (b) configurable (default local `disable`, prod `require`) | **(b)**. Cuesta una entrada en `config.go` ahora y evita partirlo en F3. |
| **D10** | Casing de la instancia Evolution | `.env.example` dice `Jairo`; el código Node y el plan dicen `GLAUK` | Necesito el **nombre canónico exacto** de la instancia (los nombres de instancia viajan en la URL y son sensibles a mayúsculas). Lo unifico en un solo valor documentado. |
| **D11** | Fallback de tabla `perfiles_glaukov` → `perfiles` | (a) conservarlo · (b) eliminarlo | **(b)** previa confirmación de que `perfiles_glaukov` es la canónica. El fallback oculta un error de esquema y duplica el trabajo en `store/socios.go`. |
| **D12** | Idioma de identificadores y Doc Comments | (a) identificadores en inglés + comentarios en español · (b) todo en español · (c) todo en inglés | **(a)**: `Config`, `NewPool`, `SendMedia` en inglés (idiomático y compatible con linters) y doc comments en español (coherente con el equipo). |
| **D13** | Quién sirve el dashboard | (a) Go sirve `public/` y `/flags` (paridad total) · (b) Node sigue sirviendo el frontend y Go solo la API | **(b) durante F1–F5**, y (a) en F5 si se busca retirar Node por completo. Definir ahora evita duplicar el mapeo de estáticos. |
| **D14** | Control de versiones | No hay repositorio git en la carpeta (verificado) | **Inicializar git (o respaldo íntegro) antes de escribir la primera línea.** Sin control de versiones no se puede revisar un diff por archivo, que es precisamente la forma de auditar el límite de 60 líneas. |
| **D15** | Ubicación del renderer Node durante la migración | (a) queda en la raíz como hoy, expuesto en `:3090` · (b) se mueve dentro de `tasashub-go/` | **(a)**: mantiene intacto el pipeline actual mientras Go se prueba. `RendererURL` en `config.go` permite apuntar a `localhost` o a otro host. |

---

## 11. Hallazgos bloqueantes (independientes de la aprobación de esta V2)

**11.1 Go no está instalado.** `go version` no resuelve en la máquina (verificado). Es prerequisito del paso 0 y sin él no hay verificación posible de ninguno de los pasos 1–9.

**11.2 Clave de API expuesta.** `.env.example:10` contiene una clave de 32 hex con pinta de real. Debe **rotarse** (no basta borrarla: ya está en un archivo versionable) y sustituirse por un placeholder del tipo `AUTHENTICATION_API_KEY=tu_api_key_aqui`. Aplicar lo mismo a `JID_GRUPO_PRUEBA` (el grupo real como valor de ejemplo invita a disparar envíos por accidente).

**11.3 Sin control de versiones.** No hay `.git` en la carpeta. Se recomienda `git init` + commit inicial **antes** de la migración (ver D14).

**11.4 Conflicto de puerto.** `Dockerfile` expone `3002` y `server.js` escucha ahí; cualquier arranque simultáneo del binario Go fallaría (ver D3).

**11.5 `client.go` del plan excede el límite** en cuanto se añaden timeout, envoltura de errores y chequeo de status (ver §3.1). Corregido por el reparto de V2.

**11.6 Discrepancia `PLAN_FASE1.md` vs `DIAGNOSTICO_MIGRACION.md`** sobre la lista de archivos (ver §3.5). Esta V2 es el documento que la resuelve; conviene anotar en `PLAN_FASE1.md` que el reparto de archivos pasa a ser el de aquí.

---

## 12. Criterio de aceptación de la Fase 1

Se acepta la Fase 1 cuando **todo** lo siguiente es verdad, medido y no estimado:

1. `go build ./...` y `go vet ./...` sin salida.
2. `gofmt -l .` no imprime archivos.
3. `go test ./...` en verde, con el caso "modo simulación no toca la red" y el caso "un solo header `apikey`".
4. `scripts/check-file-length.ps1` reporta los 12 archivos de código entre 30 y 60 líneas (excepción justificada: `client.go`, que está completo).
5. `GET /health` devuelve `status`, `service`, `instance` y `db`, y `instance` cambia con `EVOLUTION_INSTANCE` **sin recompilar**.
6. El proceso falla al arrancar, con mensaje accionable, si la base de datos no responde.
7. `.env.example` no contiene ningún secreto real.
8. Existe `git` (o respaldo) con un commit inicial previo a la migración.

---

*Este documento es una propuesta de arquitectura. No sustituye a `PLAN_FASE1.md` en sus reglas ni en su alcance funcional: sustituye únicamente su reparto de archivos, por evidenciarse incompatible con el límite de 30–60 líneas.*
