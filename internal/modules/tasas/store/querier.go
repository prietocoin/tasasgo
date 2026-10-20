// Package store es la única puerta a la base de datos de tasas: traduce filas de
// PostgreSQL a los tipos del dominio, y ninguna otra capa escribe SQL.
//
// No conoce HTTP, ni el cálculo, ni el renderizado. Recibe el contexto y el
// ejecutor de consultas por parámetro, de modo que no exista un pool global
// escondido y que las pruebas puedan sustituirlo sin tocar una base real.
package store

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Querier es el contrato mínimo que el store necesita del motor de base de datos.
//
// Se declara como interfaz, en lugar de recibir *pgxpool.Pool, por dos motivos: el
// paquete dice explícitamente qué operaciones usa (solo dos), y las pruebas pueden
// inyectar un doble sin una base de datos real. Se dejan fuera transacciones y
// lotes de sentencias a propósito: el store ejecuta una sentencia por llamada, y
// quien orqueste varias deberá decidirlo de forma visible en otra capa.
type Querier interface {
	// Query ejecuta una consulta que devuelve filas.
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	// Exec ejecuta una sentencia que no devuelve filas.
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

// Garantía en tiempo de compilación: el pool real debe seguir cumpliendo el
// contrato. Si pgx cambia una firma, esto deja de compilar aquí y no en producción.
var _ Querier = (*pgxpool.Pool)(nil)
