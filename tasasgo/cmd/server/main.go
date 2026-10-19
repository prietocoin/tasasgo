// Package main arranca el servicio HTTP de TasasGo.
//
// Este paquete solo ensambla piezas: lee la configuración, abre la base de datos
// y monta el servidor. Toda la lógica de negocio vive en internal/, para que sea
// posible probarla sin arrancar un proceso ni abrir un puerto.
package main

import (
	"context"
	"log"

	// Embebe la base de zonas horarias dentro del binario. La imagen final es
	// scratch y no lleva /usr/share/zoneinfo: sin esta importación, resolver
	// America/Caracas falla y toda hora local cae en silencio a UTC.
	_ "time/tzdata"

	"github.com/gofiber/fiber/v2"

	"tasasgo/internal/config"
	"tasasgo/internal/db"
)

// main arranca el servicio y no retorna.
//
// Si la base de datos no responde, aborta el proceso en lugar de arrancar en
// estado degradado: un fallo al iniciar se atribuye solo, mientras que el mismo
// fallo en la primera petición de un usuario llega sin contexto sobre su origen.
func main() {
	cfg := config.Load()

	pool, err := db.NewPool(context.Background(), cfg)
	if err != nil {
		log.Fatalf("base de datos no disponible: %v", err)
	}
	defer pool.Close()

	listen(newApp(cfg, pool), cfg)
}

// listen publica el puerto y registra la identidad activa del proceso.
//
// El log incluye la instancia de Evolution y el modo simulación porque son los
// dos valores que más se confunden al desplegar: verlos al arrancar ahorra
// diagnosticar en caliente por qué los mensajes salen —o no salen— por la
// instancia equivocada.
func listen(app *fiber.App, cfg *config.Config) {
	log.Printf("TasasGo ⚡ escuchando en :%s", cfg.Port)
	log.Printf("instancia Evolution: %s · simulación: %t", cfg.EvolutionInstance, cfg.SimulationMode())
	log.Fatal(app.Listen(":" + cfg.Port))
}
