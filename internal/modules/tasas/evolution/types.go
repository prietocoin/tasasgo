// Package evolution encapsula el cliente HTTP de Evolution API v2.
//
// Los tipos de wire viven en este archivo, separados de la lógica de red, para
// que cualquier otro paquete pueda importar el contrato sin arrastrar net/http.
package evolution

// SendMediaRequest describe el cuerpo de POST /message/sendMedia/{instance}.
//
// Media es base64 puro, SIN el prefijo "data:image/png;base64,". El código muerto
// que existía en Node sí lo incluía, y por eso su payload era incompatible con el
// que realmente circulaba en producción. El prefijo se elimina en el borde que
// construye la imagen, nunca aquí.
type SendMediaRequest struct {
	Number    string `json:"number"`
	MediaType string `json:"mediatype"`
	MimeType  string `json:"mimetype"`
	FileName  string `json:"fileName"`
	Media     string `json:"media"`
	Caption   string `json:"caption"`
}

// SendMediaResponse conserva la respuesta de forma permisiva.
//
// Solo Status es un contrato estable. Key y Message quedan como mapas genéricos
// hasta que exista una necesidad real de leer el identificador del mensaje: fijar
// ahora una estructura detallada obligaría a mantener campos que nadie consume.
type SendMediaResponse struct {
	Key     map[string]any `json:"key"`
	Message map[string]any `json:"message"`
	Status  string         `json:"status"`
}

// SimulationStatus es el estado sintético que devuelve el cliente cuando no hay
// EVOLUTION_API_URL configurada. Existe para poder distinguir un envío simulado
// de uno real sin lanzar un error que obligue a tratar la simulación como fallo.
const SimulationStatus = "simulated"
