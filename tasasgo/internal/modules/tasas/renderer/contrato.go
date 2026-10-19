// Package renderer convierte una cartelera ya calculada en la imagen PNG que se
// publica: delega en el servicio Node existente (Satori + Resvg), que se mantiene
// como sidecar HTTP para no reimplementar el dibujado ni perder fidelidad visual.
//
// El paquete no conoce Fiber, ni la base de datos, ni el calculador: solo el
// contrato, la traducción del dominio al JSON que espera el sidecar y el cliente
// HTTP que lo invoca. El día que el renderizado pase a Go, cambia la
// implementación de Renderer y nada más.
package renderer

import (
	"context"
	"fmt"

	"tasasgo/internal/modules/tasas/dominio"
)

// Renderer es el contrato que consumen los handlers: recibida una cartelera,
// devuelve sus bytes PNG. Se declara como interfaz para que la ruta pueda probarse
// con un doble en memoria y para que la decisión "renderiza Node o renderiza Go"
// no se filtre al resto del servicio.
type Renderer interface {
	// Render devuelve el PNG de la cartelera. Respeta el contexto: si quien llama
	// cancela o agota su plazo, la espera termina.
	Render(ctx context.Context, cartelera dominio.Cartelera) ([]byte, error)
}

// StatusError describe un sidecar que respondió con un código distinto de 200. Se
// tipa, en lugar de devolver un error plano, para distinguir "el renderer falló"
// de "no se pudo llegar al renderer", y para conservar el fragmento de respuesta
// que explica el fallo.
type StatusError struct {
	// Code es el código HTTP que devolvió el sidecar.
	Code int
	// Body conserva el inicio de la respuesta, suficiente para diagnosticar.
	Body string
}

// Error describe el fallo incluyendo el código HTTP recibido.
func (e *StatusError) Error() string {
	return fmt.Sprintf("renderer respondió HTTP %d: %s", e.Code, e.Body)
}
