// Command evoprobe valida contra la instancia real de Evolution API que el
// header "apikey" —y solo ese— alcanza para enviar media.
//
// Existe como herramienta de línea de comandos y no como ruta HTTP para que la
// validación no amplíe la superficie pública del servicio: se ejecuta una vez,
// imprime la respuesta y termina. Es la comprobación del hallazgo más incierto
// del diagnóstico, el triple header de autenticación del código Node.
package main

import (
	"context"
	"encoding/json"
	"log"

	"tasasgo/internal/config"
	"tasasgo/internal/modules/tasas/evolution"
)

// probePixel es un PNG de 1x1 transparente: el payload más pequeño posible para
// probar el envío sin depender del renderer ni de archivos locales.
const probePixel = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8AAAwAB/AGt6QAAAABJRU5ErkJggg=="

// main aborta con un mensaje explícito si falta cualquier valor necesario.
//
// Se exige JID_GRUPO_PRUEBA en lugar de usar un destino por defecto, porque un
// sondeo nunca debe poder escribir en un grupo de producción por descuido.
func main() {
	cfg := config.Load()
	if cfg.SimulationMode() {
		log.Fatal("EVOLUTION_API_URL está vacío: no hay instancia que sondear")
	}
	if cfg.JIDGrupoPrueba == "" {
		log.Fatal("JID_GRUPO_PRUEBA está vacío: se exige un destino de prueba explícito")
	}

	client := evolution.New(cfg.EvolutionAPIURL, cfg.EvolutionAPIKey, cfg.EvolutionInstance)
	request := evolution.SendMediaRequest{
		Number: cfg.JIDGrupoPrueba, MediaType: "image", MimeType: "image/png",
		FileName: "probe.png", Media: probePixel, Caption: "🧪 Sondeo TasasGo",
	}

	response, err := client.SendMedia(context.Background(), request)
	if err != nil {
		log.Fatalf("sondeo fallido: %v", err)
	}
	pretty, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		log.Fatalf("no se pudo formatear la respuesta: %v", err)
	}
	log.Printf("respuesta de la instancia %s:\n%s", cfg.EvolutionInstance, pretty)
}
