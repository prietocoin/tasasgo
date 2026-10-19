// Sidecar de renderizado: expone por HTTP el renderer Node que hasta ahora solo
// se usaba en proceso (generarCarteleraPNG). Existe para que la Fase 4 de TasasGo
// pueda pedir el PNG sin reimplementar Satori + Resvg ni la plantilla.
//
// Debe arrancarse con el directorio de trabajo en la raíz de este proyecto: el
// renderer carga la fuente y los assets con rutas relativas (assets/fonts, ...).
//
//   npm run renderer        # escucha en el puerto 3090 (RENDERER_PORT)
//
// Contrato: POST /render recibe el JSON de la cartelera y responde 200 image/png;
// si el renderizado falla, responde 500 con el mensaje en texto plano.
import express from 'express';
import { generarCarteleraPNG } from './src/modules/tasas/services/renderer.service.js';

const PORT = Number(process.env.RENDERER_PORT || 3090);
const app = express();

app.use(express.json({ limit: '10mb' }));

app.post('/render', async (req, res) => {
  try {
    const png = await generarCarteleraPNG(req.body);
    res.type('png').send(png);
  } catch (err) {
    console.error('[renderer] fallo al renderizar:', err.message);
    res.status(500).type('text/plain').send(err.message);
  }
});

app.listen(PORT, '0.0.0.0', () => {
  console.log(`[TasasHub Renderer 🎨] Sidecar escuchando en el puerto ${PORT}`);
});
