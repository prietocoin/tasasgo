package store

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"tasasgo/internal/modules/tasas/dominio"
)

// Columnas de perfiles_glaukov que el store conoce, en el orden en que scanSocio las
// lee: la lista y el escaneo viven juntos porque separarlos es como se desincronizan.
const (
	sqlSocioByName = `SELECT nombre, id_grupo, moneda_base, monedas, activo, modo_prueba, rol FROM perfiles_glaukov WHERE LOWER(nombre) = LOWER($1)`
	sqlAllSocios   = `SELECT nombre, id_grupo, moneda_base, monedas, activo, modo_prueba, rol FROM perfiles_glaukov ORDER BY nombre ASC`
)

// SocioByName busca un socio por nombre sin distinguir mayúsculas.
//
// La comparación se resuelve en la base, con LOWER, y no trayendo la tabla a memoria.
func SocioByName(ctx context.Context, q Querier, nombre string) (Socio, error) {
	return queryOne(ctx, q, sqlSocioByName, []any{nombre}, scanSocio)
}

// AllSocios devuelve todos los socios ordenados por nombre, que es el orden en que
// el panel los muestra y el que evita que la lista cambie de orden entre recargas.
func AllSocios(ctx context.Context, q Querier) ([]Socio, error) {
	return queryMany(ctx, q, sqlAllSocios, nil, scanSocio)
}

// scanSocio traduce una fila de socio, incluida la configuración de monedas en JSON.
//
// El rol se devuelve tal como está en la tabla: aplicar el valor por defecto SOCIO es
// una decisión de presentación y vive en la capa que arma la respuesta.
func scanSocio(rows pgx.Rows) (Socio, error) {
	var nombre string
	var idGrupo, monedaBase, rol *string
	var activo, modoPrueba *bool
	var raw []byte
	if err := rows.Scan(&nombre, &idGrupo, &monedaBase, &raw, &activo, &modoPrueba, &rol); err != nil {
		return Socio{}, fmt.Errorf("escaneando el socio: %w", err)
	}
	monedas, err := decodeJSON[map[string]*dominio.Moneda](raw, "monedas")
	if err != nil {
		return Socio{}, err
	}
	return Socio{
		Perfil: dominio.Perfil{
			Nombre:     nombre,
			IDGrupo:    text(idGrupo),
			MonedaBase: text(monedaBase),
			Monedas:    monedas,
		},
		Activo:     activo,
		ModoPrueba: modoPrueba,
		Rol:        text(rol),
	}, nil
}
