// Package config centraliza la lectura de variables de entorno del servicio.
//
// Es la única fuente de verdad sobre la configuración: ningún otro paquete debe
// llamar a os.Getenv ni a godotenv, para que exista un solo lugar donde buscar
// de dónde proviene un valor.
package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// loadDotEnv carga un archivo .env local antes de resolver la configuración.
//
// Solo actúa fuera de producción: en producción las variables deben venir del
// sistema, para que un .env olvidado dentro de una imagen no tape una
// configuración incompleta con valores de desarrollo. El error se ignora a
// propósito, porque la ausencia del archivo es un caso válido.
func loadDotEnv() {
	if os.Getenv("APP_ENV") == "production" {
		return
	}
	_ = godotenv.Load()
}

// getEnv devuelve el valor de la variable o el fallback cuando está ausente o
// vacía. Tratar la cadena vacía como "no definida" evita que una línea sin valor
// en el .env borre un valor por defecto de Config.
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// getEnvInt convierte la variable a entero y cae al fallback si falta o si el
// valor no es numérico. No falla a propósito: un puerto mal escrito no debe
// impedir el arranque, solo provocar el uso del valor por defecto.
func getEnvInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return fallback
}
