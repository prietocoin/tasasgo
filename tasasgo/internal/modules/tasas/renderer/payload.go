package renderer

// Tarjeta es una fila de la cartelera tal como la espera la plantilla Node y tal
// como la devuelve /calcular.
//
// Se declara aquí, y no reutilizando dominio.TarjetaPais, para que un cambio en el
// vocabulario interno no altere sin querer el JSON que cruza al sidecar: el
// compilador no protege un contrato HTTP.
//
// Compra y Venta son punteros porque el cálculo puede producir NaN —un defecto
// latente del original, ver calculator/precision.go— y JSON no lo admite: el
// servicio Node habría emitido null, y eso es lo que emite este tipo.
type Tarjeta struct {
	Code       string   `json:"code"`
	Pais       string   `json:"pais"`
	NombrePais string   `json:"nombre_pais"`
	Bandera    string   `json:"bandera"`
	Compra     *float64 `json:"compra"`
	Venta      *float64 `json:"venta"`
	Trend      string   `json:"trend"`
}

// Payload es a la vez el cuerpo exacto del POST al sidecar y el objeto que devuelve
// /calcular. Los nombres JSON son literalmente los que desestructura
// cartelera.template.js: cambiarlos rompe la imagen en silencio, sin error de
// compilación.
//
// `id_grupo` y `moneda_base` no los lee la plantilla, pero se mantienen porque el
// JSON normalizado del origen los incluía y el sidecar podría necesitarlos.
type Payload struct {
	NombreSocio string    `json:"nombre_socio"`
	IDGrupo     string    `json:"id_grupo,omitempty"`
	MonedaBase  string    `json:"moneda_base"`
	LoteTasa    string    `json:"lote_tasa"`
	Fecha       string    `json:"fecha_actualizacion"`
	Hora        string    `json:"hora_actualizacion"`
	Tarjetas    []Tarjeta `json:"tarjetas_paises"`
}
