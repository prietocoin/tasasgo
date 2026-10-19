package renderer

import "math"

// orElse devuelve el primer valor no vacío, o cadena vacía si todos lo están.
//
// Es el equivalente del operador `||` con el que el servicio original completaba
// los campos del payload: allí `t.bandera || t.code?.toLowerCase()` sustituía una
// bandera ausente por el código en minúsculas, y esa regla debe sobrevivir al
// puerto porque la imagen depende de ella.
func orElse(valores ...string) string {
	for _, v := range valores {
		if v != "" {
			return v
		}
	}
	return ""
}

// tasa traduce un número del cálculo al payload.
//
// El puntero no es un capricho: el cálculo reproduce un defecto del original —una
// tasa del orden de 1e-9 produce NaN, ver calculator/precision.go— y JSON no admite
// ese valor. El servicio Node emitía null porque JSON.stringify lo convierte; Go en
// cambio fallaría al serializar. Convirtiéndolo aquí, tanto la respuesta de
// /calcular como el cuerpo que recibe el sidecar dicen lo que decía el original.
func tasa(valor float64) *float64 {
	if math.IsNaN(valor) || math.IsInf(valor, 0) {
		return nil
	}
	return &valor
}
