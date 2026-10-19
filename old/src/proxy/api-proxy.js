import { env } from '../config/env.js';

// Cabeceras que nunca se reenvían. Las "hop-by-hop" pertenecen a este salto, no al
// siguiente: copiarlas hace que el cliente negocie una conexión que ya no existe.
// content-length se recalcula, y accept-encoding se descarta a propósito para que
// el comportamiento sea determinista: fetch descomprime lo que reciba y aquí se
// sirve el cuerpo ya en claro, así que content-encoding tampoco puede viajar.
// Dejar pasar esa cabecera haría que el navegador intentara descomprimir datos
// que ya vienen descomprimidos.
const NO_SE_REENVIA = new Set([
  'host',
  'connection',
  'keep-alive',
  'transfer-encoding',
  'content-length',
  'content-encoding',
  'accept-encoding',
  'upgrade',
  'te',
  'trailer',
  'proxy-authenticate',
  'proxy-authorization',
]);

// El cuerpo se acumula en memoria antes de reenviarlo, así que la puerta tiene que
// quedar igual de cerrada que con express.json: el mismo tope de 10 MB.
const MAX_CUERPO = 10 * 1024 * 1024;

// leerCuerpo devuelve el cuerpo crudo sin interpretarlo. Se reenvía byte a byte en
// lugar de volver a serializar req.body, para que el backend Go reciba exactamente
// lo que envió el panel y no una versión reinterpretada del JSON.
async function leerCuerpo(peticion) {
  const trozos = [];
  let total = 0;
  for await (const trozo of peticion) {
    total += trozo.length;
    if (total > MAX_CUERPO) {
      const error = new Error(`el cuerpo supera los ${MAX_CUERPO} bytes`);
      error.status = 413;
      throw error;
    }
    trozos.push(trozo);
  }
  return trozos.length > 0 ? Buffer.concat(trozos) : undefined;
}

// cabecerasCortas filtra las cabeceras de la petición de entrada.
function cabecerasCortas(cabeceras) {
  const salida = {};
  for (const [nombre, valor] of Object.entries(cabeceras)) {
    if (!NO_SE_REENVIA.has(nombre.toLowerCase())) {
      salida[nombre] = valor;
    }
  }
  return salida;
}

// apiProxy reenvía al backend Go todo lo que empiece por /api. Existe porque Node
// ya no es backend: atiende el frontend y las imágenes, y la API y sus datos
// viven en TasasGo. Mantener aquí una segunda implementación de /api fue lo que
// dejó al panel intentando abrir su propia conexión a PostgreSQL.
//
// No impone tiempo de espera: operaciones legítimamente lentas (publicar un lote
// de socios) deben poder terminar. El tope de duración corresponde al proxy
// inverso, que sí conoce el tamaño del despliegue.
export function apiProxy() {
  // Se quita la barra final para no producir rutas con doble barra, y el prefijo
  // del panel se conserva usando la URL original completa.
  const base = env.API_URL.replace(/\/+$/, '');
  return async (peticion, respuesta) => {
    const destino = base + peticion.originalUrl;
    let respuestaGo;
    try {
      respuestaGo = await fetch(destino, {
        method: peticion.method,
        headers: cabecerasCortas(peticion.headers),
        body: await leerCuerpo(peticion),
      });
    } catch (error) {
      // Un 413 es culpa de la petición; cualquier otro fallo aquí significa que el
      // backend no respondió. Se responde 502 con el destino en el cuerpo en vez
      // de dejar caer la petición del panel con un error de red opaco.
      const cuerpoGrande = error.status === 413;
      console.error(`[TasasHub ⚡] proxy hacia ${destino} falló: ${error.message}`);
      respuesta
        .status(cuerpoGrande ? 413 : 502)
        .json({
          error: cuerpoGrande
            ? 'cuerpo de la petición demasiado grande'
            : 'backend de tasas no disponible',
          destino,
          detalle: error.message,
        });
      return;
    }
    respuesta.status(respuestaGo.status);
    for (const [nombre, valor] of respuestaGo.headers) {
      if (!NO_SE_REENVIA.has(nombre.toLowerCase())) {
        respuesta.setHeader(nombre, valor);
      }
    }
    respuesta.end(Buffer.from(await respuestaGo.arrayBuffer()));
  };
}
