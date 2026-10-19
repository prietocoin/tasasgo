package evolution

import (
	"net/http"
	"time"
)

// requestTimeout acota cada llamada a Evolution API. Sin este límite, una
// instancia caída deja bloqueado al worker que publica carteleras, y el síntoma
// aparece como "el sistema se quedó colgado" en lugar de como un error de red.
const requestTimeout = 30 * time.Second

// Client es un cliente HTTP sin estado para una única instancia de Evolution API.
//
// Guarda una sola credencial y envía un solo header de autenticación: el
// diagnóstico encontró que el helper de Node mandaba tres headers simultáneos
// (apikey, Global-Api-Key y Authorization), lo que ocultaba cuál aceptaba en
// realidad la instancia.
type Client struct {
	baseURL  string
	apiKey   string
	instance string
	http     *http.Client
}

// New construye un cliente ligado a una instancia concreta.
//
// Un baseURL vacío activa el modo simulación, que permite operar el servicio
// completo sin WhatsApp: es lo que hace posible levantar y probar todo el flujo
// en un entorno sin credenciales ni red.
func New(baseURL, apiKey, instance string) *Client {
	return &Client{
		baseURL:  baseURL,
		apiKey:   apiKey,
		instance: instance,
		http:     &http.Client{Timeout: requestTimeout},
	}
}

// Simulated indica si el cliente evita la red por completo.
//
// Los llamadores lo usan para registrar el envío como intención en lugar de
// reportarlo como fallo, y para no contar como error la ausencia de credenciales
// en desarrollo.
func (c *Client) Simulated() bool {
	return c.baseURL == ""
}
