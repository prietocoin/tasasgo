# TasasGo ⚡

Servicio HTTP en Go (Fiber v2) que reemplaza de forma incremental el backend Node/Express de
TasasHub: la Fase 1 dejó el esqueleto (configuración, base de datos, cliente de Evolution API y
`/health`), la Fase 2 portó el calculador, la Fase 3 la capa de datos, la Fase 4 el renderer, la
Fase 5 los doce endpoints de `/api/v1/tasas` con su concurrencia, la Fase 6 el cliente del origen
HOO y el middleware de seguridad, y la Fase 7 la imagen Docker y el entorno orquestado. Hasta el
final de la migración el frontend y el renderer PNG siguen siendo del proyecto Node.

> **Documento de arquitectura:** `REPLANTEO_ARQUITECTURA_V2.md`
> **Plan de origen:** `PLAN_FASE1.md` · **Diagnóstico:** `DIAGNOSTICO_MIGRACION.md`

---

## Estado actual

| Fase | Contenido | Estado |
|---|---|---|
| **F1** | Esqueleto: config, db, cliente Evolution, `/health` | ✅ Implementada |
| **F2** | Calculador puro (`calculator/`) validado contra fixtures dorados | ✅ Implementada |
| **F3** | Acceso a datos (`store/`) con el pool inyectado | ✅ Implementada |
| **F4** | Renderer tras interfaz (`renderer/`) + cliente del sidecar Node | ✅ Implementada |
| **F5** | Los 12 endpoints de `/api/v1/tasas` + concurrencia | ✅ Implementada |
| **F6** | Cliente HOO + middleware de seguridad | ✅ Implementada |
| **F7** | Docker multi-stage + tzdata | ✅ Implementada |

El frontend y el renderer PNG **siguen siendo servidos por el proyecto Node** durante toda la
migración.

---

## Requisitos

- **Go 1.27 o superior** (`go version`)
- **PostgreSQL** accesible (mismas credenciales que usa hoy el servicio Node)
- **Node.js** solo si vas a levantar el renderer (Fase 4+)
- **Docker y Compose v2** solo si vas a levantar el entorno completo de una vez (ver
  *Docker: el entorno completo*)
- Opcional: [`golangci-lint`](https://golangci-lint.run/) si quieres análisis estático adicional a `go vet`

---

## Arranque rápido

```powershell
# 1. Situarse en el módulo
cd tasasgo

# 2. Crear el archivo de entorno local a partir de la plantilla
Copy-Item .env.example .env

# 3. Editar .env con las credenciales reales de PostgreSQL.
#    Deja EVOLUTION_API_URL vacío para trabajar en modo simulación.

# 4. Descargar dependencias
go mod tidy

# 5. Compilar y ejecutar
go build ./...
go run ./cmd/server
```

Verificación:

```powershell
curl http://localhost:3003/health
# {"status":"OK","service":"TasasHub Engine ⚡","instance":"Glauk","database":"ok",...}
```

Comprobar que la instancia es configurable **sin recompilar**:

```powershell
$env:EVOLUTION_INSTANCE="Glauk"; go run ./cmd/server
# El campo "instance" del /health refleja el valor activo.
```

### Modo simulación

Si `EVOLUTION_API_URL` está vacío, el cliente **no toca la red**: registra la intención y devuelve
`{"status":"simulated"}`. Esto permite levantar y probar todo el servicio sin credenciales de
WhatsApp. El campo `simulated` de `/health` expone el modo activo.

### Sondeo contra Evolution real

`cmd/evoprobe` es una herramienta de un solo disparo que envía un PNG de 1×1 píxel a
`JID_GRUPO_PRUEBA` para confirmar que el header `apikey` —y solo ese— es suficiente. Existe como CLI
y no como ruta HTTP para no ampliar la superficie pública del servicio.

```powershell
$env:EVOLUTION_API_URL="https://tu-instancia.evolution.host"
$env:EVOLUTION_API_KEY="tu_api_key"
$env:JID_GRUPO_PRUEBA="120363XXXXXXXXXX@g.us"
go run ./cmd/evoprobe
```

---

## Estructura de carpetas

```text
tasasgo/
├── go.mod                                  Definición del módulo y dependencias
├── Dockerfile                              Imagen scratch multi-stage, sin root
├── .dockerignore                           Lo que no entra al contexto de compilación
├── .env.example                            Plantilla de variables de entorno (sin secretos)
├── README.md                               Este documento
├── scripts/
│   └── check-file-length.ps1               Auditoría del límite de líneas por archivo
├── cmd/
│   ├── server/
│   │   ├── main.go                         Arranque: ensambla config, db y servidor
│   │   ├── app.go                          Construcción del servidor Fiber y manejo de errores
│   │   ├── routes.go                       Tabla de rutas y respuesta 404
│   │   └── health.go                       Contrato del endpoint /health
│   └── evoprobe/
│       └── main.go                         CLI de sondeo contra Evolution real
└── internal/
    ├── config/
    │   ├── config.go                       Qué ajustes existen y sus valores por defecto
    │   └── lookup.go                       Cómo se leen y convierten las variables de entorno
    ├── db/
    │   └── postgres.go                     Ciclo de vida del pool de conexiones
    ├── middleware/
    │   └── security.go                     CORS y la política de marcos del iframe
    └── modules/
        └── tasas/
            ├── dependencias.go             Deps: lo que los handlers comparten, inyectado
            ├── routes.go                   Las trece rutas bajo /api/v1/tasas
            ├── respuesta.go                Formas de respuesta JSON y lectura del cuerpo
            ├── respuesta_render.go         Las respuestas en texto plano del endpoint PNG
            ├── cartelera_actual.go         Pipeline compartido: socio y lote vigente
            ├── envio.go                    Renderizar y publicar una cartelera por WhatsApp
            ├── concurrencia.go             Pool de 5 workers con plazo por socio
            ├── servicio_publicar.go        Publicar el lote vigente a todos los socios
            ├── handler_lectura.go          GET /monitor y GET /socios
            ├── handler_socio.go            PUT /socios/:socio
            ├── handler_lote.go             POST /guardar-lote y DELETE /lote/:id_tasa
            ├── handler_render.go           GET /calcular/:socio y GET|POST /render
            ├── handler_disparo.go          POST /disparar
            ├── handler_publicar.go         POST /publicar-lote y POST /disparar-lote
            ├── handler_remesas.go          POST /notificar-remesas
            ├── handler_hoo.go              GET /hoo-borrador
            ├── routes_test.go              Suite de rutas con dobles en memoria
            ├── dominio/
            │   └── tipos.go                Vocabulario compartido del negocio de tasas
            ├── calculator/
            │   ├── precision.go            Regla de precisión de tasas (1e-8)
            │   ├── number.go               Normalización de números heterogéneos
            │   ├── catalogo.go             Banderas, nombres de país y divisas de referencia
            │   ├── base.go                 Moneda nativa del socio (denominador del cálculo)
            │   ├── triangulacion.go        Cross-rate y márgenes de compra/venta
            │   ├── tendencia.go            up / down / equal frente al lote anterior
            │   ├── tiempo.go               Fecha y hora de publicación (America/Caracas)
            │   ├── tarjeta.go              Una tarjeta: incluir o descartar cada divisa
            │   ├── cartelera.go            Orquestador del cálculo completo
            │   ├── calculator_test.go      Suite contra los fixtures dorados
            │   └── testdata/
            │       ├── golden_calculator.json        Salida real del servicio Node de origen
            │       └── gen_golden_calculator.mjs     Generador de esos fixtures (Node)
            ├── store/
            │   ├── querier.go             Contrato mínimo que el store pide a la base
            │   ├── escaneo.go             Camino único de lectura de filas y error centinela
            │   ├── jsonb.go               Traducción entre columnas JSON y mapas
            │   ├── valores.go             Lectura de una columna por nombre, aunque falte
            │   ├── esquema.go             Qué columnas tiene de verdad la tabla heredada
            │   ├── lotes.go               Lote vigente, identificador de lote y su escaneo
            │   ├── lotes_escritura.go     Alta y baja de lotes
            │   ├── socios.go              Modelo de socio y su actualización
            │   ├── socios_lectura.go      Búsquedas de socios
            │   ├── socios_escaneo.go      Fila de socio a Socio, leyendo por nombre
            │   ├── socios_escritura.go    UPDATE armado con las columnas que existen
            │   ├── notificaciones.go      Registro de notificaciones de tasas
            │   ├── store_test.go          Suite con doble de base de datos
            │   └── socios_test.go         Regresión del esquema heredado sin `activo`
            ├── renderer/
            │   ├── contrato.go            Interfaz Renderer y el error de estado HTTP
            │   ├── payload.go             Forma exacta del JSON que espera el sidecar Node
            │   ├── valores.go             Valores vacíos y no finitos, y sus valores por defecto
            │   ├── normalizar.go          Cómo se traduce el dominio al payload y se serializa
            │   ├── sidecar.go             El cliente del renderer: construcción y Render
            │   ├── peticion.go            Detalle de la petición: cabeceras, estado, cuerpo
            │   └── sidecar_test.go        Suite contra un sidecar de mentira (httptest)
            ├── hoo/
            │   ├── contrato.go            Interfaz Borrador y la forma del borrador de tasas
            │   ├── cliente.go             El cliente HTTP del origen global de tasas
            │   ├── mapeo.go               Qué campo del origen va a qué campo del borrador
            │   ├── valores.go             Las reglas de coerción heredadas de JavaScript
            │   └── cliente_test.go        Suite contra un origen de mentira y el payload real
            └── evolution/
                ├── types.go                Contrato de datos con Evolution API
                ├── client.go               Construcción y estado del cliente
                ├── send.go                 La llamada HTTP POST /message/sendMedia
                └── send_test.go            Pruebas del cliente
```

### Por qué `cmd/` y `internal/`

- `cmd/` contiene **procesos ejecutables**. Cada subcarpeta es un binario.
- `internal/` contiene **bibliotecas privadas** del módulo. Go impide que código externo al módulo
  importe `internal/`, lo que convierte esta carpeta en una frontera de API verificada por el
  compilador y no por convención.
- La dirección de dependencias es siempre `cmd/ → internal/`. Ningún paquete de `internal/` conoce la
  existencia de `cmd/`. Dentro del módulo, `tasas` (los handlers) importa a `store`, `calculator`,
  `renderer`, `evolution` y `hoo`, pero ninguno de esos cinco importa a `tasas`: la gracia tiene un solo
  sentido y no hay ciclos.

---

## El calculador (Fase 2)

`internal/modules/tasas/calculator` es **lógica pura**: no lee la base de datos, no llama por red y no
renderiza. Recibe el `Perfil`, el lote vigente, el lote anterior y el instante; devuelve la
`dominio.Cartelera`. Ese instante se pasa por parámetro en lugar de leerlo del reloj para que cada
resultado sea reproducible y comparable.

El cálculo se reparte por **eje de cambio**, no por tamaño: la regla de precisión (`precision.go`)
cambiaría si cambia la política de decimales; los márgenes (`triangulacion.go`) si cambia cómo se
aplican; el catálogo (`catalogo.go`) el día que se sume una divisa. `cartelera.go` orquesta y
`tarjeta.go` decide si una divisa se publica o se descarta —ese es el único archivo que añade el
diseño original, porque la decisión de incluir/excluir no cabía con la orquestación bajo el límite de
60 líneas.

### Fixtures dorados

`calculator/testdata/golden_calculator.json` **no se escribió a mano**: se generó ejecutando el
servicio Node actual sobre 18 casos borde (catálogo completo, triangulación con moneda base no
trivial, tendencias, divisas ausentes, `activo` omitido o `false`, cero numérico frente a `"0"`,
márgenes negativos y altos, cadenas con separador de miles, claves en minúscula), más 37 valores
límite de la regla de precisión y 13 instantes horarios. Es la definición ejecutable de "el puerto se
comporta igual que el original".

```powershell
# Suite completa, con el resumen por caso
go test -v ./internal/modules/tasas/calculator/

# Regenerar los fixtures si cambia el calculador Node de origen
node internal/modules/tasas/calculator/testdata/gen_golden_calculator.mjs ../old internal/modules/tasas/calculator/testdata/golden_calculator.json
```

La comparación tolera 1e-9 relativo: absorbe la última cifra de las funciones matemáticas de Go frente
a las de V8, pero un error real de fórmula o de regla de precisión queda muy por encima de esa cota y
la prueba lo muestra con el diff campo por campo.

### Comportamientos del original que se conservan a propósito

| Comportamiento | Por qué se conserva |
|---|---|
| Una tasa del orden de `1e-9` produce **NaN** | El original redondea a cero y entra en la rama de magnitud, donde `log10(0)` es infinito. Se reproduce para no alterar ninguna cartelera durante la migración; **es un defecto latente** y está anotado en `precision.go`. |
| `99.995` se publica como `99` | El corte de la regla es `> 99.99` y redondea hacia abajo: es la política vigente, no un error de redondeo. |
| `"0"` (cadena) descarta la divisa, `0` (número) la publica con tasa 1 | El operador `||` de JavaScript distingue ambos casos. Está fijado en los fixtures. |
| Fechas como `17-oct` o `01-ene` | Es el formato real del locale `es-VE`, con guion, minúsculas y `sept` para septiembre. `tiempo.go` declara esas abreviaturas como dato en lugar de depender del idioma del sistema. |
| Zona horaria con desplazamiento fijo `-4` | Venezuela no aplica horario de verano desde 2016 y así la imagen final no necesita empaquetar `tzdata`. Si algún día se publicaran fechas históricas, es el único archivo que cambia. |

### Una desviación consciente

Las tarjetas salen **ordenadas por código de divisa**. El original las recorría en el orden de
inserción del JSON del perfil, que es un detalle de almacenamiento y no una decisión de negocio. Con
el orden alfabético, el mismo lote produce siempre exactamente la misma imagen (y la comparación con
los fixtures es posible). El contenido de cada tarjeta es idéntico.

## La capa de datos (Fase 3)

`internal/modules/tasas/store` es **el único paquete que escribe SQL**. No conoce HTTP ni el cálculo:
recibe un `context.Context` y un ejecutor de consultas por parámetro, devuelve tipos del `dominio` y no
guarda ningún pool global. Los fallos del motor viajan siempre envueltos con `%w`, para no perder la
causa original.

El contrato con el motor es la interfaz `Querier` (dos operaciones: `Query` y `Exec`). Se declara como
interfaz, y no como `*pgxpool.Pool`, por dos motivos: deja escrito qué operaciones usa el paquete, y
permite inyectar un doble en las pruebas. El pool real la satisface sin adaptadores, y una afirmación de
compilación (`var _ Querier = (*pgxpool.Pool)(nil)`) obliga a que siga siendo así: si pgx cambia una
firma, deja de compilar el paquete en lugar de fallar en producción.

### Cómo se leen las columnas

Las columnas `jsonb` se escanean como **bytes crudos** y se interpretan en `jsonb.go`. Así funciona
igual si la columna es `jsonb` o `text` —el servicio original tenía que defenderse de las dos con
`typeof x === 'string'`— y convierte el caso difícil en una función pura, verificable sin base de datos.
Las columnas de texto opcionales (`id_tasa`, `id_grupo`, `moneda_base`, `rol`) se leen como anulables:
un nulo se traduce a cadena vacía, que es lo que el resto del sistema ya sabe interpretar.

### Pruebas sin base de datos

`store_test.go` usa `pgxmock` y compara el SQL **exacto**, no una expresión regular: una consulta
distinta rompe la prueba en lugar de pasar desapercibida. Ninguna prueba abre una conexión —el entorno
de desarrollo no alcanza la base de producción, y una prueba que depende de la red deja de ser una
prueba.

La suite se verificó por no vacuidad: cambiando `LIMIT 2` por `LIMIT 1`, devolviendo `nil` en lugar de
`ErrNotFound` y alterando el relleno de `NextLoteID`, los tres cambios fueron detectados con el detalle
del caso.

### Decisiones que se apartan del código original

| Decisión | Motivo |
|---|---|
| El lote vigente se lee con `ORDER BY id DESC` | Es lo que ejecuta el servicio vivo. El archivo muerto `db.service.js` lo hacía por `created_at`; dos definiciones de "lote vigente" publican el lote equivocado. |
| `SELECT *` en socios y mapeo por nombre de columna | El esquema heredado no es el que suponíamos: `perfiles_glaukov` no tiene `activo`, y una lista fija de columnas hizo que `/socios` fallara en producción con `42703`. Con `SELECT *` y lectura por nombre, la tabla manda y una columna ausente queda en su valor cero, igual que el `undefined` del original. La respuesta sigue sin arrastrar columnas: `Socio` expone solo el contrato del panel. |
| `db.service.js` **no** se porta | No lo importa nadie: es código muerto, y portar código muerto es migrar defectos que nadie probó. |
| `UpdateSocio` devuelve error si el socio no existe | El original respondía `success:true` con un socio vacío y el panel mostraba "guardado" sin haber guardado nada. |
| El `rol` es opcional en `SocioUpdate` y el `UPDATE` lo escribe con `COALESCE(..., rol)` | El panel **no envía** `rol`, pero el `UPDATE` original hacía `rol = $4` con `rol \|\| 'SOCIO'`: cada guardado de monedas reseteaba el rol del socio a SOCIO. Con el rol como puntero y el `COALESCE`, un rol ausente no se escribe y la pérdida silenciosa desaparece. |

### El esquema heredado manda

La tabla real de socios no tiene todas las columnas que el código Node aparentaba usar: en
producción, `perfiles_glaukov` no tiene `activo`, y la primera versión del port la enumeraba en el
`SELECT`, así que `/api/v1/tasas/socios` respondía `42703`. Las lecturas piden ahora `SELECT *` y
traducen por nombre de columna (`socios_escaneo.go`), de modo que lo que la tabla no tenga queda en
su valor cero, que es lo que hacía el servicio original al leer una propiedad ausente. El guardado
(`socios_escritura.go`) pregunta antes qué columnas existen —`SELECT * ... LIMIT 0`— y arma el
`UPDATE` solo con esas; si no quedara ninguna, falla en vez de responder un éxito vacío. No se
adivina el esquema y no se cae a la tabla `perfiles`.

### Nota sobre la versión de pgx

`go.mod` fija `github.com/jackc/pgx/v5` en **v5.10.0** a propósito. La v5.11.0 añadió `Rows.TypeMap` y
su propio CHANGELOG advierte que "las implementaciones propias de `Rows`, incluidas las de simulación,
deben añadir ese método": `pgxmock`, que es el doble que usan estas pruebas, todavía no lo hace.
Escribir a mano un doble de toda la interfaz `pgx.Rows` costaba más código que el que ahorra. La versión
se sube cuando el doble la soporte.

---

## El renderer (Fase 4)

`internal/modules/tasas/renderer` convierte una `dominio.Cartelera` ya calculada en los bytes PNG que
se publican. **No dibuja nada**: delega en el renderer Node (Satori + Resvg) que ya existe, para no
reimplementar 323 líneas de plantilla ni arriesgar una imagen distinta. Lo que Go aporta es la
frontera: una interfaz, un formato de intercambio congelado y un cliente HTTP probado sin arrancar
Node.

El contrato que consumirá la Fase 5 es una sola operación:

```go
type Renderer interface {
    Render(ctx context.Context, cartelera dominio.Cartelera) ([]byte, error)
}
```

### El contrato por el cable

El cliente hace `POST {RENDERER_URL}/render` con `Content-Type: application/json` y el cuerpo que
construye `normalizar.go`: `nombre_socio`, `id_grupo`, `moneda_base`, `lote_tasa`,
`fecha_actualizacion`, `hora_actualizacion` y `tarjetas_paises[]` con `code`, `pais`, `nombre_pais`,
`bandera`, `compra`, `venta` y `trend`. Son literalmente los nombres que desestructura
`cartelera.template.js`. **HTTP 200 con los bytes PNG en el cuerpo es el único éxito**; cualquier
otro código llega tipado como `StatusError` (con el código y el inicio de la respuesta), y cualquier
fallo de red o de plazo como error envuelto con `%w`.

`payload.go` declara los tipos del intercambio aparte de `dominio` a propósito: el compilador no
protege un contrato HTTP, así que un cambio en el vocabulario interno no debe poder alterar el JSON
que cruza al sidecar sin que alguien lo decida.

### Normalización

`normalizar.go` reproduce el helper `normalizarParaRenderer` del controlador legado, incluidos sus
valores por defecto: un país sin nombre cae a su código, una tarjeta sin bandera al código en
minúsculas y una tarjeta sin tendencia a `equal`. Las filas salen de un slice no nulo, de modo que un
socio sin divisas produce `[]` y no `null`.

### El plazo

`Render` envuelve el contexto recibido con `context.WithTimeout` (30 s por defecto). El plazo
efectivo es el más corto de los dos: quien llama puede imponer su presupuesto y el cliente sigue
protegido de una espera infinita. El plazo es un campo del `Sidecar`, así que la prueba lo baja a
20 ms en lugar de esperar medio minuto.

### Pruebas sin Node

`sidecar_test.go` levanta un `httptest.Server` y cubre los tres casos que importan: respuesta
correcta (comprobando además el JSON que recibe el servidor, no solo los bytes que devuelve), error
500 convertido en `StatusError`, y plazo agotado. Ninguna prueba necesita el sidecar real.

La suite se verificó por no vacuidad: quitando el valor por defecto de `trend`, invirtiendo la
comprobación de estado y eliminando el `WithTimeout`, los tres cambios fueron detectados con el
mensaje del caso correspondiente.

### Por qué cinco archivos y no tres

El plan de arquitectura preveía `contrato.go`, `normalizar.go` y `sidecar.go`. Con el límite de
30–60 líneas, añadir el payload y el detalle de la petición HTTP a esos tres los dejaba por encima
de 60. Se separaron por eje de cambio: el contrato y su error, la forma del JSON, la traducción del
dominio, el cliente, y el detalle de una petición. Ninguno de los cinco supera las 60 líneas.

### El sidecar Node

El renderer Node no exponía ningún endpoint: `generarCarteleraPNG` se llamaba en proceso desde el
controlador. Para poder consumirlo desde Go se añadió `old/renderer-server.js`, un envoltorio mínimo
que publica ese mismo renderer por HTTP y se arranca con:

```bash
npm run renderer        # en old/ : escucha en el 3090 (RENDERER_PORT)
```

Es el único archivo nuevo del proyecto Node y no toca nada existente: importa la misma función y la
expone en `POST /render`. Debe arrancarse desde la raíz de `old/`, porque el renderer carga la fuente
y los assets con rutas relativas al directorio de trabajo.

### Verificación de fidelidad (PNG byte-idéntico)

La prueba se hizo con el sidecar en marcha y el mismo payload por los dos caminos: el Node de
siempre llamando a `generarCarteleraPNG` en proceso, y el cliente Go haciendo `POST /render`. Ambos
produjeron exactamente el mismo archivo:

| Camino | Bytes | SHA-256 |
|---|---:|---|
| Node en proceso | 859 676 | `9b15577a…674883bb` |
| Go → sidecar HTTP | 859 676 | `9b15577a…674883bb` |

El PNG que sirve Go es **idéntico byte a byte** al que sirve Node hoy, que era el criterio de salida
de la Fase 4. La comparación debe repetirse si cambian la plantilla o los assets.

---

## La capa HTTP (Fase 5)

`internal/modules/tasas` es la capa de entrada: traduce peticiones de Fiber a llamadas del store, del
calculador y del renderer, y devuelve las respuestas con la forma que publica el servicio Node. **No
escribe SQL, no calcula tasas y no dibuja**; recibe esas piezas ya construidas y las conecta.

### Las trece rutas

El contrato es el de `tasas.routes.js`, con los mismos métodos y el mismo orden de segmentos:

| Método y ruta | Handler | Qué hace |
|---|---|---|
| `GET /monitor` | `Monitor` | El lote vigente para el panel |
| `GET /socios` | `Socios` | La lista de socios con el rol por defecto aplicado |
| `PUT /socios/:socio` | `GuardarSocio` | Guarda monedas y los campos del panel que existan en la tabla |
| `GET /hoo-borrador` | `BorradorHoo` | Borrador del origen externo (HOO) |
| `POST /guardar-lote` | `GuardarLote` | Guarda un lote nuevo sin publicarlo |
| `DELETE /lote/:id_tasa` | `EliminarLote` | Borra un lote |
| `POST /publicar-lote` | `PublicarLote` | Publica el lote vigente a todos los socios activos |
| `POST /disparar-lote` | `PublicarLote` | **Alias** de la anterior, como en el original |
| `POST /disparar` | `Disparar` | Publica la cartelera de un solo socio |
| `GET /calcular/:socio` | `Calcular` | La cartelera ya normalizada, en JSON |
| `GET /render/:socio` | `Render` | La cartelera en PNG |
| `POST /render` | `Render` | Igual, con el socio en el cuerpo |
| `POST /notificar-remesas` | `NotificarRemesas` | Registra la notificación del lote vigente |

`routes.go` solo declara rutas: no decide de dónde salen las dependencias. Eso ocurre en
`cmd/server/routes.go`, que compone el pool, el sidecar, el cliente de Evolution y la configuración,
y llama a `tasas.Register(app, deps)`.

### Dependencias inyectadas

`Deps` (`dependencias.go`) reúne el `Querier`, el `Renderer`, el cliente de Evolution, el `Borrador` de
HOO, la `Config` y un reloj `Now` que las pruebas sustituyen. Los handlers son **métodos** de `Deps`, así
que ninguna función sale a buscar sus dependencias y **no existe estado global**. `Now` es una función
y no `time.Now` directo porque la fecha publicada forma parte de la respuesta.

### El pipeline compartido

`cartelera_actual.go` contiene la única definición del cálculo de una cartelera: buscar el socio, leer
los dos últimos lotes y calcular. El controlador original repetía esos pasos en cuatro endpoints; aquí
`/calcular`, `/render` y `/disparar` comparten el mismo camino, de modo que la definición de "lote
vigente" (`loteVigente`, con el lote anterior igual al vigente cuando solo hay uno) no puede divergir
entre endpoints.

`renderer.Payload` y `renderer.Normalizar` se hicieron públicos en esta fase precisamente para eso:
`/calcular` devuelve **el mismo objeto** que se le manda al sidecar, no una copia que pueda
desincronizarse.

### Concurrencia

`concurrencia.go` ejecuta la publicación con un **pool de 5 workers** y un plazo de **30 s por socio**.
El original publicaba en serie dentro de un `for`: una instancia de WhatsApp lenta retrasaba a todos los
demás. Los socios con `activo === false` se descartan **antes** de publicar y no cuentan como
procesados, igual que el original. Si la tabla no tiene la columna `activo`, el interruptor llega
nulo y no se descarta a nadie: el original comparaba contra `false` exacto, así que una propiedad
ausente tampoco apagaba a nadie.

Como el orden en que terminan los workers es azaroso, el array `detalle` se **ordena por nombre de
socio** antes de responder. Sin ese orden, dos publicaciones idénticas podrían devolver la lista en
orden distinto y el panel mostraría un cambio donde no lo hay.

### Pruebas sin base, sin red y sin Node

`routes_test.go` monta la aplicación completa y la interroga con `app.Test`, sin abrir un puerto:

- **store**: `pgxmock` como doble del pool (el mismo que usa la Fase 3);
- **renderer**: un `rendererFalso` que devuelve un PNG mínimo;
- **evolution**: el cliente real en **modo simulación** (sin URL), que no toca la red;
- **hoo**: se inyecta `nil` para ejercitar la rama defensiva del handler, que responde un error
  explícito en lugar de caerse.

Cubre el contrato JSON de `/monitor`, la imagen y el **404 en texto plano** de `/render`, y que
`/disparar-lote` (el alias) devuelva el detalle ordenado y sin los socios apagados. La suite se verificó
por **no vacuidad**: invirtiendo el orden del `sort` y cambiando el 404 de `/render` por un 200, las dos
mutaciones fueron detectadas con el mensaje del caso.

### Desviaciones conscientes de esta fase

| Desviación | Motivo |
|---|---|
| Un pool de 5 workers en vez del `for` secuencial | Una instancia lenta ya no bloquea a los demás socios. El conjunto de resultados es el mismo; el orden se restituye ordenando por nombre. |
| `detalle` ordenado alfabéticamente | El orden del pool es azaroso y la respuesta debe ser reproducible. |
| Una tasa `NaN` se publica como `null` | JSON no admite `NaN`. El original emitía `null` al stringificar; Go fallaría al serializar, así que la conversión es explícita en `renderer/valores.go`. No se toca el calculador, que sigue reproduciendo el defecto. |
| `/render` responde **texto plano** (`Socio no encontrado`, `Error generando PNG`) | Paridad estricta con el original: su cliente es un navegador que pinta la imagen, no el panel. Es el único endpoint que no habla JSON y vive aislado en `respuesta_render.go`. |
| `notificacion` devuelve solo `id_tasa` | Es lo único que el store puede afirmar sin conocer el resto del esquema (ver Fase 3). |
| `POST /disparar` lee el socio dos veces | Prefiere la lectura de más antes que duplicar el pipeline de la cartelera. Es un endpoint administrativo, no un camino caliente. |
| El sello `CREATED_AT` usa `...T..:..:...000Z` | Es exactamente lo que produce `new Date().toISOString()`; con el RFC3339 de Go cambiaría de forma. |
| Un `id_tasa` vacío se responde como `""` y no se omite | El original habría omitido la clave (`undefined`). La diferencia solo aparece con un lote sin identificador, que el calculador ya publica como `T000`. |

### Por qué diecisiete archivos

El plan preveía diecisiete archivos; la entrega son **dieciséis en el paquete `tasas` más
`hoo/contrato.go`**, que son exactamente diecisiete. Dos ajustes los explica el propio límite de 30
líneas: `errores.go` se integró en `respuesta.go` (las respuestas de error son la misma decisión que
las de éxito y por separado quedaban por debajo del mínimo) y `handler_calculo.go` se integró en
`handler_render.go` (`/calcular` y `/render` son las dos caras de la misma cartelera). En sentido
contrario se añadió `respuesta_render.go`, porque el **texto plano** de `/render` es una convención de
respuesta con su propio motivo para cambiar y merecía un archivo donde verse entera.

A esos diecisiete se suman dos archivos que no son handlers: `routes_test.go` (la suite de rutas) y
`renderer/valores.go`, que separa las conversiones de valores del payload para poder expresar la regla
`NaN → null` sin que `normalizar.go` pase de 60 líneas. Además se modificaron tres archivos ya
existentes: `renderer/payload.go` y `renderer/normalizar.go` (DTO con `*float64` y `Normalizar` público),
`store/socios.go` (el `COALESCE` del rol), `cmd/server/routes.go` (el montaje) y sus dos suites
(`store_test.go`, `sidecar_test.go`).

---

## El origen HOO y la seguridad (Fase 6)

### El cliente del origen externo

`internal/modules/tasas/hoo` consume el origen público de tasas globales y traduce su respuesta al
borrador que el panel muestra antes de guardar un lote. No hay credencial: la API es abierta, y el
original tampoco enviaba ninguna cabecera.

El paquete se reparte por eje de cambio: `cliente.go` habla HTTP (con un plazo propio de 30 s),
`mapeo.go` decide **qué campo del origen va a qué campo del borrador**, y `valores.go` guarda las
**reglas de coerción heredadas de JavaScript**. `Mapear` es una función pura —entran bytes, sale el
borrador—, así que se prueba sin red.

Dos rarezas del original que se conservan y quedan fijadas en las pruebas:

- El cuerpo puede ser un **objeto o un array**: el original tomaba `data[0]` sin preguntar, y eso es
  lo que hace `primero`.
- El `||` de JavaScript trata al **cero como ausente**: `Number(item.usd || 1.00)` publica `1`
  cuando el origen manda `usd: 0`. Por eso las cinco divisas que hoy no aparecen en la respuesta real
  (`usd`, `usdt`, `pyusd`, `ecu`, `pan`) se publican con su valor por defecto.

| Campo del origen | Campo del borrador | Conversión del original |
|---|---|---|
| `id` | `ID` | `String(x \|\| '')` |
| `ars`, `clp`, `cop`, `ves` | en mayúsculas | `String(x \|\| 0)` |
| `bob`, `brl`, `cad`, `crc`, `dop`, `eur`, `mxn`, `pen`, `pyg` | en mayúsculas | `Number(x \|\| 0)` |
| `usd`, `usdt`, `ecu`, `pan` | en mayúsculas | `Number(x \|\| 1.00)` |
| `pyusd` | `PYUSD` | `Number(x \|\| 1.20)` |
| `created_at` | `CREATED_AT` | tal cual, o el instante actual |

Un texto que no sea un número produce `null` (en el original, `NaN`, que `JSON.stringify` convertía a
`null`). Los campos numéricos del DTO son punteros precisamente para poder declararlo sin romper la
serialización.

### Verificación contra el origen real

La prueba se hizo contra la API en vivo, y el JSON normalizado es **exactamente** el que produciría
el controlador Node para el mismo payload:

```json
{"ID":"83600","ARS":"1618","BOB":11.85,"BRL":5.06,"CAD":1.45,"CLP":"993","COP":"3162","CRC":455,
 "DOP":62.4,"EUR":0.9,"MXN":18.56,"PEN":3.45,"PYG":5765,"VES":"1006","USD":1,"USDT":1,
 "PYUSD":1.2,"ECU":1,"PAN":1,"CREATED_AT":"2026-10-19T02:35:35.030Z"}
```

Ese payload quedó congelado en `cliente_test.go`, de modo que la fidelidad se revisa en cada
`go test` sin depender de que HOO esté en pie.

### Middleware de seguridad

`internal/middleware/security.go` replica lo que `server.js` hacía en Express: CORS abierto y una
política de marcos que permite embeber el panel en el laboratorio.

| Cabecera | Valor | Por qué |
|---|---|---|
| `Content-Security-Policy` | `frame-ancestors 'self' https://lab.jairokov.com` | Solo el propio origen y el laboratorio pueden mostrar el panel en un iframe. |
| `X-Frame-Options` | **eliminada** | Su `DENY` por defecto anularía la política. Nada en Fiber la añade, así que la eliminación es defensiva: si alguna capa la pone, se quita. |
| `Access-Control-Allow-Origin` | `*` | El panel se sirve desde otro origen mientras Node y Go coexisten. |

Se monta en `newApp` **antes** de las rutas, que es cuando Fiber permite que un middleware afecte a
todo lo que venga después. La prueba lo comprueba poniendo `X-Frame-Options` aguas arriba: sin eso,
borrarla o no borrarla daría exactamente el mismo resultado observable.

---

## Docker: el entorno completo (Fase 7)

Un solo comando levanta base de datos, renderer y servicio Go desde la **raíz del repositorio**:

```bash
docker compose up --build
```

```text
Tasashub/                     raíz del repositorio
├── .gitignore                Protege el .env que Compose lee para sustituir variables
├── docker-compose.yml        Orquesta los tres servicios y su red interna
├── old/                      Proyecto Node (frontend y renderer PNG)
│   ├── Dockerfile            Se reutiliza cambiando el comando a npm run renderer
│   └── .dockerignore         Excluye node_modules del contexto de compilación
└── tasasgo/                  Servicio Go
    ├── Dockerfile            Imagen scratch multi-stage, sin root
    └── .dockerignore
```

### Qué se levanta

| Servicio | Origen | Puerto publicado | Rol |
|---|---|---|---|
| `postgres` | `postgres:16-alpine` | `127.0.0.1:5432` | Base `automat`, con el volumen `datos-postgres`. |
| `renderer-sidecar` | `./old` + `npm run renderer` | `127.0.0.1:3090` | Renderer PNG en Node (Satori + Resvg). |
| `tasasgo` | `./tasasgo` | `127.0.0.1:3003` | El servicio Go. |

Los tres puertos se publican **solo en `127.0.0.1`**: es un entorno de desarrollo, con contraseña
conocida y sin TLS, así que no se expone a la red local. Para abrirlo a la red, quita el prefijo en
`docker-compose.yml`.

### Cómo se hablan entre sí

Los servicios se referencian por **nombre de servicio**, nunca por `localhost`: dentro de un
contenedor, `localhost` es el propio contenedor. De ahí `DB_HOST=postgres` y
`RENDERER_URL=http://renderer-sidecar:3090`.

`tasasgo` espera a que Postgres esté **sano** (`condition: service_healthy`, con una sonda
`pg_isready`). Sin eso el arranque en frío se caería: el servicio aborta si la base no responde al
iniciar, y en los primeros segundos Postgres todavía está inicializando su directorio de datos.

### La base arranca vacía

El volumen empieza sin tablas: `/health` responde igual (`"database":"ok"`), pero los endpoints
fallan hasta que cargues el esquema. Desde un respaldo:

```bash
docker compose exec -T postgres psql -U postgres -d automat < respaldo.sql
```

### El frontend no forma parte del stack

El panel web sigue en Node y no se levanta aquí, así que `docker compose up` no publica el puerto
3002. Si lo necesitas mientras pruebas, arráncalo aparte (`cd old && npm start`): el stack **sí**
expone Postgres en `127.0.0.1:5432` justamente para que ese proceso lo alcance.

Ese panel ya no consulta la base de datos: reenvía todo `/api` al servicio Go con el proxy
`old/src/proxy/api-proxy.js`, y toma el destino de `API_URL` (`http://localhost:3003` si no está
definida). Con eso Node queda como servidor de frontend e imágenes, que es el estado final buscado.

### Comprobar que quedó en pie

```bash
curl http://localhost:3003/health
# {"status":"OK","service":"TasasHub Engine ⚡","instance":"Glauk","database":"ok","simulated":true,...}
```

Si `EVOLUTION_API_URL` quedó vacío —el caso por defecto—, `"simulated":true` y **no se envía nada a
ningún grupo real**. Para apuntar a WhatsApp real, define las variables en un `.env` en la raíz o en
el entorno antes de levantar el stack.

### La imagen de Go

```dockerfile
FROM golang:1.27-alpine AS builder   # compila
FROM scratch                          # ejecuta
```

- **Binario estático**: `CGO_ENABLED=0` y `-ldflags="-s -w"`. El resultado se verificó como ELF
  x86-64 **sin intérprete dinámico** y sin referencia a glibc, que es lo que `scratch` exige.
- **Sin root**: el proceso corre como `10001:10001`. Al no haber shell, tampoco hay escalada local
  que intentar. El precio es que no puedes `docker exec` para depurar; el diagnóstico se hace por
  logs.
- **Zona horaria**: `America/Caracas` viaja **embebida en el binario** (`_ "time/tzdata"` en
  `cmd/server/main.go`), no depende del sistema operativo. En `scratch` no existe
  `/usr/share/zoneinfo`: sin ese import, resolver la zona falla y las horas caen a UTC **en
  silencio**.
- **Certificados TLS**: se copia el bundle de `ca-certificates` de la etapa de compilación. Sin él,
  toda llamada HTTPS a Evolution y a HOO fallaría.
- **Multi-arquitectura**: `GOOS` y `GOARCH` salen de `TARGETOS`/`TARGETARCH`, con `linux/amd64` por
  defecto, así que el mismo Dockerfile compila en un Mac ARM sin editarlo.

### Por qué el renderer sigue siendo Node

`renderer-sidecar` reutiliza el `old/Dockerfile` cambiando el comando. Reescribir la plantilla de la
cartelera en Go cambiaría el PNG byte a byte: la fidelidad verificada en la Fase 4 (mismo SHA-256)
depende de Satori y Resvg, no de nosotros.

`old/.dockerignore` no es cosmético: el Dockerfile de Node hace `COPY . .` **después** de
`npm install`, así que sin él los `node_modules` del equipo (1 088 archivos, con binarios nativos de
Windows) se copiarían encima de las dependencias Linux recién instaladas.

---

## Reglas de arquitectura (NO NEGOCIABLES)

Estas tres reglas se verifican con herramientas, no con buena voluntad.

### 1. Responsabilidad Única (SRP) estricta

**Cada archivo tiene un único motivo para cambiar.** El criterio de corte es el *eje de cambio*, no
la cantidad de líneas:

- Si dos funciones cambiarían por razones distintas, van en archivos distintos.
- Ejemplos vigentes: `client.go` (cómo se construye el cliente) ≠ `send.go` (el protocolo HTTP) ≠
  `types.go` (el contrato de datos).
- Los handlers se separan **por recurso HTTP**, nunca por capa: `handler_lote.go`, `handler_disparo.go`,
  `handler_lectura.go`. Un `handler.go` único para 12 endpoints es imposible bajo el límite de líneas.
- Las tablas de datos (banderas, nombres de país, mapeos) son **datos**, no lógica: van en archivos
  propios como `catalogo.go`.

### 2. Límite estricto de 30–60 líneas por archivo

| Categoría | Rango permitido |
|---|---|
| Archivos `*.go` de producción | **30 – 60 líneas** |
| Archivos `*_test.go` | 0 – 100 líneas (techo ampliado para alojar tablas de casos) |
| `go.mod`, `.env.example`, `Dockerfile`, `README.md` | Sin límite (no son código) |
| `testdata/*` (fixtures y su generador) | Sin límite (datos de prueba, no código del servicio) |
| `scripts/*` | Referencia: ≤ 60 líneas |

**Se cuentan TODAS las líneas del archivo**, incluidas las vacías, los Doc Comments, el `package`,
los `import` y las llaves de cierre. La regla aplica al archivo completo, no al "código útil".

- Menos de 30 líneas solo se acepta si el archivo está **completo** (no si le falta lógica).
- Más de 60 líneas **está prohibido**. Si un archivo crece, se parte por eje de cambio; nunca se
  reduce la documentación para hacerlo caber.

Auditoría:

```powershell
.\scripts\check-file-length.ps1
# Imprime la tabla Archivo / Lineas / Min / Max / Estado y falla con código 1
# si algún archivo está fuera de rango.
```

### 3. Doc Comments obligatorios

Todo `package`, todo `type` y toda función exportada lleva un comentario que explica **la intención
(el porqué)**, no la mecánica (el qué). El nombre ya describe el qué.

- Los comentarios se escriben **en español**; los identificadores, **en inglés**.
- Los campos privados y las funciones internas se documentan cuando su intención no es evidente por
  el nombre.
- Un comentario que solo repite el nombre de la función se considera ausente.

### Verificación completa antes de dar algo por terminado

```powershell
gofmt -l .                 # Debe imprimir vacío (formato canónico)
go vet ./...               # Debe salir limpio
go build ./...             # Debe compilar
go test ./...              # Debe pasar
.\scripts\check-file-length.ps1
```

### Cumplimiento medido al cerrar la Fase 7

| Regla | Cómo se comprueba | Resultado |
|---|---|---|
| SRP: un archivo, una razón de cambio | Revisión por eje de cambio; un paquete se parte cuando el motivo para cambiar difiere | 63 archivos `.go`, cada uno con un único motivo |
| Límite 30–60 líneas en producción y ≤100 en pruebas | `.\scripts\check-file-length.ps1` | exit 0, los 63 dentro de rango |
| Doc Comments con la intención (el porqué) | Comentario en todo `package`, `type` y función exportada | Sin pendientes |
| Sin estado global | `Config` y `Deps` se inyectan por parámetro | Ningún `os.Getenv` fuera de `internal/config` |
| Formato y análisis estático | `gofmt -l .` · `go vet ./...` · `go build ./...` | Los tres, limpios |
| Pruebas | `go test ./...` | 21 pruebas en 7 paquetes, ninguna falla |

---

## Configuración

Todas las variables se leen **una sola vez** en `internal/config`. Ningún otro paquete llama a
`os.Getenv`: si necesitas un valor de entorno nuevo, se añade a `Config` y se inyecta por parámetro.

| Variable | Default | Descripción |
|---|---|---|
| `PORT` | `3003` | Puerto HTTP del servicio Go. Node usa `3002` durante la coexistencia. |
| `APP_ENV` | `development` | Si vale `production`, **no** se carga el archivo `.env`. |
| `DB_HOST` | `localhost` | Host de PostgreSQL. |
| `DB_PORT` | `5432` | Puerto de PostgreSQL. |
| `DB_USER` | `postgres` | Usuario. |
| `DB_PASSWORD` | *(vacío)* | Contraseña. Sin fallback: requerida en producción. |
| `DB_NAME` | `automat` | Base de datos. |
| `DB_SSLMODE` | `disable` | `disable` en local, `require` en producción. |
| `EVOLUTION_API_URL` | *(vacío)* | Base de Evolution API. **Vacío = modo simulación.** |
| `EVOLUTION_API_KEY` | *(vacío)* | Credencial única. |
| `EVOLUTION_INSTANCE` | `Glauk` | Instancia de WhatsApp. Configurable sin recompilar. |
| `JID_GRUPO_PRUEBA` | *(vacío)* | Destino para envíos de prueba. Sin default a propósito. |
| `RENDERER_URL` | `http://localhost:3090` | Sidecar Node del renderer PNG (`npm run renderer`). |
| `HOO_GLOBAL_RATES_URL` | `https://hoo.jairokov.com/global-rates` | Origen público de tasas globales. API abierta, sin credencial. |

### Sobre el archivo `.env`

`godotenv` se carga **solo fuera de producción** (`APP_ENV != production`). En producción las
variables deben venir del sistema, para que un `.env` olvidado dentro de una imagen no tape una
configuración incompleta con valores de desarrollo.

---

## Decisiones de diseño heredadas del diagnóstico

| Decisión | Motivo |
|---|---|
| Un **solo** header de autenticación (`apikey`) | El helper Node enviaba `apikey`, `Global-Api-Key` y `Authorization` a la vez: redundancia que ocultaba cuál aceptaba realmente la instancia. |
| `Media` es base64 **puro**, sin prefijo `data:` | El código muerto `evolution.service.js` incluía el prefijo y su payload era incompatible con el que realmente se usaba en producción. |
| Un único `EvolutionClient` | No hay servicio paralelo ni código muerto: una sola implementación del envío. |
| Verificación de `StatusCode` antes de decodificar | Un 401 de Evolution responde HTML; sin validar el estado, el error se reportaría como "fallo de JSON" y ocultaría la causa real. |
| Timeout de 30 s por petición | Una instancia caída no debe bloquear indefinidamente a quien publica carteleras. |
| `Ping()` al arrancar (fail-fast) | Un error de configuración de base de datos aparece al arrancar, no en la primera petición de un usuario. |
| Sin fallback de tabla `perfiles_glaukov` → `perfiles` | El fallback del código Node ocultaba errores de esquema y duplicaba el trabajo de la capa de datos. |

---

## Solución de problemas

| Síntoma | Causa probable | Qué hacer |
|---|---|---|
| `base de datos no disponible: ...` al arrancar | Credenciales o `DB_SSLMODE` incorrectos, o PostgreSQL apagado | Revisar `.env`, confirmar que el servidor responde |
| `/health` muestra `"database":"unreachable"` | El pool se abrió pero la conexión se perdió | Revisar red/PostgreSQL; el proceso sigue vivo a propósito |
| El servicio no arranca y no hay mensaje de configuración | `APP_ENV=production` y variables sin definir | Definir las variables en el sistema, no en `.env` |
| Puerto ocupado (`address already in use`) | Está corriendo el servicio Node en `3002`, o hay otra instancia de Go | Detener el proceso anterior o cambiar `PORT` |
| `address already in use` con el Node detenido | Una instancia previa de `go run` sigue viva | Buscar el proceso y terminarlo |

---

*Regla de oro del repositorio: si un archivo necesita superar las 60 líneas, la solución es dividirlo
por responsabilidad, nunca comprimirlo ni recortar su documentación.*
