// Package db gestiona el ciclo de vida del pool de conexiones a PostgreSQL.
//
// No contiene consultas: la capa de acceso a datos llega en la Fase 3, cuando
// exista el paquete store/ con una consulta por caso de uso.
package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"tasasgo/internal/config"
)

// pingTimeout acota la verificación de arranque. Es corto a propósito: si la
// base no responde en este plazo, arrancar en estado degradado solo traslada el
// problema a la primera petición de un usuario.
const pingTimeout = 5 * time.Second

// NewPool abre el pool y verifica la conexión antes de devolver el control.
//
// Verificar aquí convierte un error de configuración en un fallo inmediato y
// legible al arrancar, en lugar de un error diferido y difícil de atribuir.
func NewPool(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dsn(cfg))
	if err != nil {
		return nil, fmt.Errorf("abrir pool postgres: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, pingTimeout)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("verificar conexión postgres: %w", err)
	}
	return pool, nil
}

// dsn construye la cadena de conexión en el formato que exige pgx.
//
// Se mantiene separada de NewPool para poder verificar el formato sin abrir un
// socket, y para que el día que aparezca una URL completa de proveedor exista un
// único lugar donde decidir cuál de las dos formas tiene prioridad.
func dsn(cfg *config.Config) string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode,
	)
}
