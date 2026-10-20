package tasas

import (
	"github.com/gofiber/fiber/v2"
)

// BorradorHoo responde GET /hoo-borrador con los valores de cambio del día.
//
// El borrador es la propuesta que el panel muestra para que el usuario la revise y
// la corrija antes de guardar un lote: este endpoint no toca la base ni publica
// nada, solo consulta el origen externo y lo traduce.
//
// El origen se consume a través de la interfaz hoo.Borrador, de modo que el handler
// no depende de un cliente concreto y puede probarse con un doble en memoria: el
// cliente HTTP vive en hoo/cliente.go y el mapeo de los veinte campos, en hoo/mapeo.go.
//
// La dependencia puede ser nula —el cableado la deja opcional a propósito— y en ese
// caso se responde un error explícito en lugar de provocar un pánico: un fallo con
// mensaje se diagnostica, y un pánico tira el proceso entero, dejando sin carteleras
// a todos los socios y no solo a quien pidió el borrador.
func (d *Deps) BorradorHoo(c *fiber.Ctx) error {
	if d.Hoo == nil {
		return fallo(c, fiber.StatusInternalServerError, "origen HOO no configurado")
	}
	cambios, err := d.Hoo.Obtener(c.Context())
	if err != nil {
		return errorInterno(c, err)
	}
	return ok(c, fiber.Map{"borrador": cambios})
}
