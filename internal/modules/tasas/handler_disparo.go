package tasas

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"tasasgo/internal/modules/tasas/store"
)

// Disparar responde POST /disparar: publica la cartelera de un solo socio.
//
// El modo prueba se toma del cuerpo y, si no viene, del que el socio tenga guardado;
// el original repetía esa decisión justo antes de elegir el grupo de destino.
func (d *Deps) Disparar(c *fiber.Ctx) error {
	var cuerpo struct {
		Socio      string `json:"socio"`
		ModoPrueba *bool  `json:"modo_prueba"`
	}
	if err := parsearCuerpo(c, &cuerpo); err != nil {
		return fallo(c, fiber.StatusBadRequest, "cuerpo de disparo inválido")
	}
	socio, err := store.SocioByName(c.Context(), d.Pool, cuerpo.Socio)
	if errors.Is(err, store.ErrNotFound) {
		return noEncontrado(c, "Socio no encontrado")
	}
	if err != nil {
		return errorInterno(c, err)
	}
	cartelera, err := d.carteleraDe(c.Context(), cuerpo.Socio)
	if err != nil {
		return errorInterno(c, err)
	}
	respuesta, err := d.enviarCartelera(c.Context(), socio, cartelera, modoPruebaDe(cuerpo.ModoPrueba, socio))
	if err != nil {
		return errorInterno(c, err)
	}
	return ok(c, fiber.Map{"socio": socio.Nombre, "evolutionResponse": respuesta})
}

// modoPruebaDe decide el modo prueba de un disparo: manda el del cuerpo, si viene, y
// si no el que el socio tenga guardado.
func modoPruebaDe(enviado *bool, socio store.Socio) bool {
	if enviado != nil {
		return *enviado
	}
	return socio.ModoPrueba != nil && *socio.ModoPrueba
}
