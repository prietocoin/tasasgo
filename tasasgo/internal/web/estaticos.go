// Package web entrega el panel desde el propio servicio, sin depender del
// proyecto Node que lo servía.
//
// El dashboard es un HTML con su manifiesto, su service worker y las banderas:
// nada de eso necesita un proceso aparte, y mantenerlo era la última razón para
// conservar el contenedor de Node. El paquete no conoce la API, ni la base de
// datos, ni el cálculo; su única decisión es qué se sirve y con qué prioridad.
package web

import "github.com/gofiber/fiber/v2"

// indice es el archivo que responde en la raíz. El panel es una sola página: las
// vistas se alternan por JavaScript, así que un "/" tiene que devolver este HTML
// y no un listado del directorio.
const indice = "index.html"

// Register monta los archivos del panel en la raíz del sitio.
//
// Debe llamarse **después** de las rutas de la API: el servicio de archivos
// responde a cualquier ruta que nadie haya reclamado antes, así que montarlo
// primero taparía /api/v1/tasas. Un archivo que no existe no se responde desde
// aquí: la petición sigue su curso, de modo que el 404 en JSON del servicio
// sigue siendo el que ve el cliente.
//
// La raíz llega por parámetro porque el proceso no puede adivinar desde dónde se
// lo ejecuta: en el contenedor el directorio de trabajo es "/" y en desarrollo es
// la carpeta del módulo.
func Register(app *fiber.App, raiz string) {
	app.Static("/", raiz, fiber.Static{
		Index: indice,
		// La caché de fasthttp mantiene abiertos los descriptores de cada archivo
		// durante diez segundos: tras editar el panel, el navegador seguiría
		// viendo la versión vieja sin ninguna pista de por qué. Reabrir el
		// archivo en cada petición es lo que hacía el servicio Node.
		CacheDuration: -1,
	})
}
