package tasas

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"tasasgo/internal/modules/tasas/store"
)

// ok responde 200 con `success:true` y los campos del endpoint.
//
// El servicio Node construye cada respuesta añadiendo esa marca al mismo nivel que
// los datos, así que el panel lee siempre la misma forma venga de donde venga.
func ok(c *fiber.Ctx, campos fiber.Map) error {
	campos["success"] = true
	return c.JSON(campos)
}

// fallo responde con el código indicado y la forma `{success:false,error}`.
func fallo(c *fiber.Ctx, status int, mensaje string) error {
	return c.Status(status).JSON(fiber.Map{"success": false, "error": mensaje})
}

// noEncontrado responde 404 conservando el mensaje del endpoint.
func noEncontrado(c *fiber.Ctx, mensaje string) error {
	return fallo(c, fiber.StatusNotFound, mensaje)
}

// errorInterno responde 500 con el texto del fallo, como el original.
func errorInterno(c *fiber.Ctx, err error) error {
	return fallo(c, fiber.StatusInternalServerError, err.Error())
}

// errorDeSocio traduce el fallo de buscar o guardar un socio: ausente es 404 y
// cualquier otra causa es un fallo del servidor.
func errorDeSocio(c *fiber.Ctx, err error) error {
	if errors.Is(err, store.ErrNotFound) {
		return noEncontrado(c, "Socio no encontrado")
	}
	return errorInterno(c, err)
}

// parsearCuerpo interpreta el JSON del cuerpo cuando existe.
//
// Un cuerpo vacío no es un error: el panel llama a varios endpoints sin datos, y
// Express los trataba como un objeto vacío. Un JSON malformado sí lo es.
func parsearCuerpo(c *fiber.Ctx, destino any) error {
	if len(c.Body()) == 0 {
		return nil
	}
	return c.BodyParser(destino)
}
