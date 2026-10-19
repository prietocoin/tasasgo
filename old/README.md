# ⚡ TasasHub (tasakov)

Microservicio de alto rendimiento desacoplado de **Glaukov Engine**, diseñado para el cálculo financiero de paridades de divisas y la generación ultra rápida de carteleras visuales (PNG) en milisegundos.

---

## 🚀 Características Principales

* **Renderizado Ligero (Sin Puppeteer/Chromium):** Utiliza **Satori** + **`@resvg/resvg-js`** (compilado en Rust) para renderizar SVG a PNG en ~30ms con un consumo plano de ~30 MB de RAM.
* **Aislamiento de Recursos:** Elimina el consumo excesivo de memoria en los workers de BullMQ y protege el CORE de Glaukov de caídas por procesos huérfanos.
* **Lógica Financiera Determinista:**
  * **Cartelera Visual:** Aplica la regla fija usando $\vert{}porcentaje\vert{}$, donde el Depósito (Compra) siempre suma el margen a la tasa base y el Pago (Venta) siempre lo resta.
  * **Normalización Automática:** Casting y sanitización estricta de tipos de datos numéricos provenientes del lote base.

---

## 🗄️ Estructura de Datos Consultada

El servicio consulta exclusivamente dos tablas en PostgreSQL:

1. **`tasas_glaukov`**: Obtiene el último lote oficial publicado (`id_tasa`, p. ej. `T059`) con el mapa de paridades de divisas.
2. **`perfiles_glaukov`**: Lee la configuración individual del socio (`nombre`, `id_grupo`, `moneda_base` y el objeto JSONB `monedas` con los porcentajes $D/P$).

---

## 📂 Estructura del Proyecto

```text
tasashub/
├── assets/
│   └── fonts/
│       └── Inter-Bold.ttf
├── src/
│   ├── config/
│   │   ├── db.js
│   │   └── env.js
│   ├── utils/
│   │   └── formatters.js
│   └── modules/
│       └── tasas/
│           ├── controllers/
│           │   └── tasas.controller.js
│           ├── routes/
│           │   └── tasas.routes.js
│           ├── services/
│           │   ├── calculator.service.js
│           │   ├── db.service.js
│           │   └── renderer.service.js
│           └── templates/
│               └── cartelera.template.js
├── .env.example
├── .gitignore
├── Dockerfile
├── docker-compose.yml
├── package.json
├── README.md
└── server.js
