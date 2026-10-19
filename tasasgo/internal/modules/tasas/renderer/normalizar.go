package renderer

import (
	"encoding/json"
	"fmt"
	"strings"

	"tasasgo/internal/modules/tasas/dominio"
)

// serializar traduce la cartelera del dominio al JSON exacto que espera el sidecar
// Node. Con estos tipos la codificación no puede fallar, pero el error se propaga
// envuelto con %w en lugar de ignorarse: el día que aparezca un campo no
// serializable, el diagnóstico ya estará puesto.
func serializar(cartelera dominio.Cartelera) ([]byte, error) {
	body, err := json.Marshal(Normalizar(cartelera))
	if err != nil {
		return nil, fmt.Errorf("serializar cartelera: %w", err)
	}
	return body, nil
}

// Normalizar traduce la cartelera del dominio al payload que cruza al sidecar y que
// devuelve /calcular, con las filas ya completadas por tarjetaPara.
//
// Se exporta porque los dos usos son el mismo objeto: el cuerpo del POST y la
// respuesta del endpoint de cálculo. Duplicar el mapeo sería la forma más rápida de
// que la imagen y la API dejaran de decir lo mismo.
//
// Devuelve siempre un slice no nulo para que el JSON sea `[]` y no `null` cuando el
// socio no publica ninguna divisa.
func Normalizar(cartelera dominio.Cartelera) Payload {
	filas := make([]Tarjeta, 0, len(cartelera.Tarjetas))
	for _, t := range cartelera.Tarjetas {
		filas = append(filas, tarjetaPara(t))
	}
	return Payload{
		NombreSocio: cartelera.NombreSocio,
		IDGrupo:     cartelera.IDGrupo,
		MonedaBase:  cartelera.MonedaBase,
		LoteTasa:    cartelera.LoteTasa,
		Fecha:       cartelera.Fecha,
		Hora:        cartelera.Hora,
		Tarjetas:    filas,
	}
}

// tarjetaPara completa los tres campos que el origen toleraba vacíos —país, bandera
// y tendencia— y traduce sus dos tasas al tipo que admite JSON.
func tarjetaPara(t dominio.TarjetaPais) Tarjeta {
	return Tarjeta{
		Code:       t.Code,
		Pais:       orElse(t.NombrePais, t.Code),
		NombrePais: orElse(t.NombrePais, t.Code),
		Bandera:    orElse(t.Bandera, strings.ToLower(t.Code)),
		Compra:     tasa(t.Compra),
		Venta:      tasa(t.Venta),
		Trend:      orElse(t.Trend, "equal"),
	}
}
