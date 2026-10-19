package tasas

import (
	"context"
	"encoding/base64"
	"fmt"

	"tasasgo/internal/modules/tasas/dominio"
	"tasasgo/internal/modules/tasas/evolution"
	"tasasgo/internal/modules/tasas/store"
)

// enviarCartelera renderiza la cartelera del socio y la publica donde corresponda.
//
// Devuelve la respuesta de Evolution para que el handler pueda reportarla. La imagen
// viaja en base64 puro, sin el prefijo `data:`, y el pie distingue prueba de oficial.
func (d *Deps) enviarCartelera(ctx context.Context, socio store.Socio, cartelera dominio.Cartelera, modoPrueba bool) (*evolution.SendMediaResponse, error) {
	destino, err := d.destino(socio, modoPrueba)
	if err != nil {
		return nil, err
	}
	png, err := d.Renderer.Render(ctx, cartelera)
	if err != nil {
		return nil, err
	}
	prefijo := "🚀 [OFICIAL]"
	if modoPrueba {
		prefijo = "🧪 [PRUEBA]"
	}
	return d.Evolution.SendMedia(ctx, evolution.SendMediaRequest{
		Number:    destino,
		MediaType: "image",
		MimeType:  "image/png",
		FileName:  "cartelera.png",
		Media:     base64.StdEncoding.EncodeToString(png),
		Caption:   fmt.Sprintf("%s Cartelera *%s* — Lote: %s", prefijo, cartelera.NombreSocio, cartelera.LoteTasa),
	})
}

// destino decide el grupo al que se envía.
//
// En modo prueba el destino es el grupo de pruebas, sin importar el del socio: es el
// seguro que evita publicar precios sin revisar. Fuera de modo prueba se usa el del
// socio y, si no tiene, el de pruebas, que es el último recurso del original.
func (d *Deps) destino(socio store.Socio, modoPrueba bool) (string, error) {
	if !modoPrueba && socio.IDGrupo != "" {
		return socio.IDGrupo, nil
	}
	return d.grupoDePrueba()
}

// grupoDePrueba devuelve el JID de pruebas y falla si falta, en lugar de inventar un
// destino: publicar en un grupo equivocado es peor que no publicar, y un JID de
// producción no debe estar escrito en el código.
func (d *Deps) grupoDePrueba() (string, error) {
	if d.Config.JIDGrupoPrueba == "" {
		return "", fmt.Errorf("JID_GRUPO_PRUEBA sin configurar")
	}
	return d.Config.JIDGrupoPrueba, nil
}
