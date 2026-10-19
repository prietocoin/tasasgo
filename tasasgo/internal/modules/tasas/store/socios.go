package store

import (
	"context"

	"tasasgo/internal/modules/tasas/dominio"
)

// Socio es un perfil tal como vive en perfiles_glaukov, más los interruptores
// operativos que no participan del cálculo.
//
// Se separa de dominio.Perfil, que es la entrada del calculador, porque estos campos
// son de la fila y no del negocio del cálculo: al calculador no le importa si un
// socio está activo, y quien decide publicar necesita saberlo antes. Embebe el perfil
// para que quien tenga un Socio pueda pasárselo al calculador sin conversión.
type Socio struct {
	dominio.Perfil
	Activo     *bool  `json:"activo"`
	ModoPrueba *bool  `json:"modo_prueba"`
	Rol        string `json:"rol"`
}

// SocioUpdate son los campos que el panel puede cambiar de un socio.
//
// Activo y ModoPrueba son punteros porque la columna admite nulo y el panel envía
// nulo cuando no toca el interruptor: distinguir "ponlo en falso" de "no lo toques"
// es la diferencia entre apagar un socio a propósito y apagarlo sin querer.
//
// El rol tampoco viaja en la petición del panel, así que también es puntero: un rol
// ausente no se escribe. El original hacía `rol = $4` con `rol || 'SOCIO'`, de modo
// que cada guardado de monedas reseteaba el rol del socio; el COALESCE de la
// sentencia convierte esa pérdida silenciosa en una no-operación.
type SocioUpdate struct {
	Monedas    map[string]*dominio.Moneda
	Activo     *bool
	ModoPrueba *bool
	Rol        *string
}

// sqlUpdateSocio guarda los campos que el panel administra y devuelve la fila
// completa, para que la interfaz se pinte con lo que quedó guardado.
const sqlUpdateSocio = `UPDATE perfiles_glaukov SET monedas = $1, activo = $2, modo_prueba = $3, rol = COALESCE($4, rol) WHERE LOWER(nombre) = LOWER($5) RETURNING nombre, id_grupo, moneda_base, monedas, activo, modo_prueba, rol`

// UpdateSocio guarda la configuración de un socio y devuelve cómo quedó.
//
// Si el nombre no existe devuelve ErrNotFound. El servicio original respondía éxito
// con un socio vacío en ese caso, y el panel mostraba "guardado" sin haber guardado
// nada: silenciar ese desacuerdo es peor que devolver el error.
func UpdateSocio(ctx context.Context, q Querier, nombre string, cambios SocioUpdate) (Socio, error) {
	raw, err := encodeJSON(cambios.Monedas)
	if err != nil {
		return Socio{}, err
	}
	args := []any{raw, cambios.Activo, cambios.ModoPrueba, cambios.Rol, nombre}
	return queryOne(ctx, q, sqlUpdateSocio, args, scanSocio)
}
