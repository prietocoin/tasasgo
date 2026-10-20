package calculator

import "tasasgo/internal/modules/tasas/dominio"

// cardFor construye la tarjeta de una divisa, o decide que no debe publicarse.
//
// Devuelve false en los tres casos en que el origen omite la divisa, y los tres
// están fijados en los fixtures dorados: el socio la desactivó explícitamente, el
// lote no la cotiza y no es una divisa de referencia, o la tasa normalizada quedó
// en cero. Un cuarto caso, la configuración nula, también se descarta: en el origen
// un socio sin objeto de configuración no publica esa divisa, mientras que un
// objeto vacío sí la publica sin márgenes.
func cardFor(code string, cfg *dominio.Moneda, current, previous dominio.Lote, baseNow, baseBefore float64) (dominio.TarjetaPais, bool) {
	if cfg == nil || (cfg.Activo != nil && !*cfg.Activo) {
		return dominio.TarjetaPais{}, false
	}

	normalized := NormalizeCode(code)
	raw, quoted := current.Tasas[normalized]
	if !quoted && !IsFixedRate(normalized) {
		return dominio.TarjetaPais{}, false
	}

	rate := NormalizeNumber(rawOrFallback(raw))
	if rate == 0 {
		return dominio.TarjetaPais{}, false
	}

	cross := CrossRate(rate, baseNow)
	trend := TrendEqual
	if previousRaw, ok := previous.Tasas[normalized]; ok {
		trend = Trend(cross, CrossRate(NormalizeNumber(previousRaw), baseBefore))
	}
	buy, sell := Prices(cross, cfg.Porcentaje)

	return dominio.TarjetaPais{
		Code:       normalized,
		NombrePais: CountryName(normalized),
		Bandera:    Flag(normalized),
		Compra:     buy,
		Venta:      sell,
		Trend:      trend,
	}, true
}
