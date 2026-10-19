package tasas

import (
	"context"
	"sync"
	"time"

	"tasasgo/internal/modules/tasas/store"
)

// workers es el número máximo de publicaciones simultáneas.
//
// Se acota a cinco porque del otro lado hay una instancia de WhatsApp: disparar
// todos los socios a la vez satura Evolution y convierte una publicación correcta en
// una tanda de errores de red.
const workers = 5

// timeoutSocio acota lo que puede tardar un socio: renderizar la imagen y enviarla.
// Sin este límite, un socio lento retendría su worker indefinidamente.
const timeoutSocio = 30 * time.Second

// resultado es el estado de la publicación de un socio, con la forma que el panel
// muestra en el detalle de la respuesta.
type resultado struct {
	Socio  string `json:"socio"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

// enParalelo aplica el trabajo a cada socio con un máximo de cinco a la vez y un
// plazo independiente por socio.
//
// El original publicaba en serie: una instancia lenta retrasaba a todos los demás.
// El conjunto de resultados es el mismo, pero el orden en que terminan es azaroso,
// así que quien responda debe ordenarlo antes de mostrarlo.
func (d *Deps) enParalelo(ctx context.Context, socios []store.Socio, trabajo func(context.Context, store.Socio) resultado) []resultado {
	salida := make([]resultado, len(socios))
	limite := make(chan struct{}, workers)
	var wg sync.WaitGroup
	for i, socio := range socios {
		wg.Add(1)
		go func(i int, socio store.Socio) {
			defer wg.Done()
			limite <- struct{}{}
			defer func() { <-limite }()
			acotado, cancelar := context.WithTimeout(ctx, timeoutSocio)
			defer cancelar()
			salida[i] = trabajo(acotado, socio)
		}(i, socio)
	}
	wg.Wait()
	return salida
}
