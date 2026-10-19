import pool from '../../../config/db.js';
import { generarCarteleraPNG } from '../services/renderer.service.js';
import { calcularTasasCartelera } from '../services/calculator.service.js';

const HOO_GLOBAL_RATES_URL = 'https://hoo.jairokov.com/global-rates';
const JID_GRUPO_PRUEBA = process.env.JID_GRUPO_PRUEBA || '120363421142957552@g.us';
const EVOLUTION_API_URL = process.env.EVOLUTION_API_URL;
// Lee AUTHENTICATION_API_KEY con fallback a EVOLUTION_API_KEY
const EVOLUTION_API_KEY = process.env.AUTHENTICATION_API_KEY || process.env.EVOLUTION_API_KEY; 
const EVOLUTION_INSTANCE = process.env.EVOLUTION_INSTANCE || 'GLAUK';

// -------------------------------------------------------------
// HELPER: Normalizar la salida de calculator.service para renderer.service
// -------------------------------------------------------------
function normalizarParaRenderer(datosCalculados) {
  const tarjetas = datosCalculados.tarjetasPaises || datosCalculados.tarjetas_paises || [];
  
  return {
    ...datosCalculados,
    tarjetas_paises: tarjetas.map(t => ({
      code: t.code,
      pais: t.nombre_pais || t.pais || t.code,
      nombre_pais: t.nombre_pais || t.pais || t.code,
      bandera: t.bandera || t.code?.toLowerCase(),
      compra: t.compra,
      venta: t.venta,
      trend: t.trend || 'equal'
    }))
  };
}

// -------------------------------------------------------------
// HELPER: Enviar Imagen base64 vía Evolution API
// -------------------------------------------------------------
async function enviarMediaEvolution(jid, base64Data, caption) {
  if (!EVOLUTION_API_URL || !EVOLUTION_INSTANCE) {
    console.log(`[SIMULACIÓN EVOLUTION] Enviado a ${jid}: ${caption}`);
    return { success: true, simulated: true };
  }

  const endpoint = `${EVOLUTION_API_URL}/message/sendMedia/${EVOLUTION_INSTANCE}`;
  
  const base64Puro = base64Data.includes(',') 
    ? base64Data.split(',')[1] 
    : base64Data;

  const payload = {
    number: jid,
    mediatype: 'image',
    mimetype: 'image/png',
    fileName: 'cartelera.png',
    media: base64Puro,
    caption: caption
  };

  try {
    const response = await fetch(endpoint, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'apikey': EVOLUTION_API_KEY,
        'Global-Api-Key': EVOLUTION_API_KEY,
        'Authorization': `Bearer ${EVOLUTION_API_KEY}`
      },
      body: JSON.stringify(payload)
    });

    const resData = await response.json();
    console.log(`[EVOLUTION RES ${response.status} -> ${jid}]:`, JSON.stringify(resData));
    return resData;
  } catch (err) {
    console.error(`[EVOLUTION ERROR -> ${jid}]:`, err.message);
    throw err;
  }
}

// =============================================================
// CONTROLADORES / ENDPOINTS
// =============================================================

// 1. OBTENER MONITOR GLOBAL
export async function obtenerTasasMonitor(req, res) {
  try {
    const result = await pool.query('SELECT * FROM tasas_glaukov ORDER BY id DESC LIMIT 1');
    if (result.rows.length === 0) {
      return res.json({ success: true, id_tasa: 'T000', tasas: {} });
    }
    const row = result.rows[0];
    const tasasObj = typeof row.tasas === 'string' ? JSON.parse(row.tasas) : (row.tasas || {});
    return res.json({
      success: true,
      id_tasa: row.id_tasa || 'T000',
      tasas: tasasObj
    });
  } catch (error) {
    console.error('[Error obtenerTasasMonitor]:', error.message);
    return res.status(500).json({ success: false, error: error.message });
  }
}

// 2. LISTAR SOCIOS
export async function obtenerListaSocios(req, res) {
  try {
    let result;
    try {
      result = await pool.query('SELECT * FROM perfiles_glaukov ORDER BY nombre ASC');
    } catch (e) {
      result = await pool.query('SELECT * FROM perfiles ORDER BY nombre ASC');
    }

    const sociosNormalizados = result.rows.map(s => ({
      ...s,
      rol: s.rol || 'SOCIO',
      monedas: typeof s.monedas === 'string' ? JSON.parse(s.monedas) : (s.monedas || {})
    }));

    return res.json({ success: true, socios: sociosNormalizados });
  } catch (error) {
    console.error('[Error obtenerListaSocios]:', error.message);
    return res.status(500).json({ success: false, error: error.message });
  }
}

// 3. ACTUALIZAR SOCIO
export async function actualizarMonedasSocio(req, res) {
  try {
    const { socio } = req.params;
    const { monedas, activo, modo_prueba, rol } = req.body;

    let query = `
      UPDATE perfiles_glaukov 
      SET monedas = $1, activo = $2, modo_prueba = $3, rol = $4
      WHERE LOWER(nombre) = LOWER($5)
      RETURNING *;
    `;
    let result;
    try {
      result = await pool.query(query, [JSON.stringify(monedas), activo, modo_prueba, rol || 'SOCIO', socio]);
    } catch (e) {
      query = `UPDATE perfiles SET monedas = $1, activo = $2, modo_prueba = $3, rol = $4 WHERE LOWER(nombre) = LOWER($5) RETURNING *;`;
      result = await pool.query(query, [JSON.stringify(monedas), activo, modo_prueba, rol || 'SOCIO', socio]);
    }

    return res.json({ success: true, socio: result.rows[0] });
  } catch (error) {
    console.error('[Error actualizarMonedasSocio]:', error.message);
    return res.status(500).json({ success: false, error: error.message });
  }
}

// 4. TRAER BORRADOR DESDE HOO
export async function traerBorradorHoo(req, res) {
  try {
    const response = await fetch(HOO_GLOBAL_RATES_URL);
    if (!response.ok) throw new Error(`HOO API HTTP Status ${response.status}`);

    const data = await response.json();
    const item = Array.isArray(data) ? data[0] : data;

    const borradorMapeado = {
      ID: String(item.id || ''),
      ARS: String(item.ars || 0),
      BOB: Number(item.bob || 0),
      BRL: Number(item.brl || 0),
      CAD: Number(item.cad || 0),
      CLP: String(item.clp || 0),
      COP: String(item.cop || 0),
      CRC: Number(item.crc || 0),
      DOP: Number(item.dop || 0),
      EUR: Number(item.eur || 0),
      MXN: Number(item.mxn || 0),
      PEN: Number(item.pen || 0),
      PYG: Number(item.pyg || 0),
      VES: String(item.ves || 0),
      USD: Number(item.usd || 1.00),
      USDT: Number(item.usdt || 1.00),
      PYUSD: Number(item.pyusd || 1.20),
      ECU: Number(item.ecu || 1.00),
      PAN: Number(item.pan || 1.00),
      CREATED_AT: item.created_at || new Date().toISOString()
    };

    return res.json({ success: true, borrador: borradorMapeado });
  } catch (error) {
    console.error('[Error Traer Borrador HOO]:', error.message);
    return res.status(500).json({ success: false, error: error.message });
  }
}

// 5. SOLO GUARDAR EN BD
export async function guardarLoteSolo(req, res) {
  try {
    const { tasas } = req.body;

    if (!tasas || typeof tasas !== 'object') {
      return res.status(400).json({ success: false, error: 'Estructura de tasas inválida.' });
    }

    const lastRow = await pool.query('SELECT id_tasa FROM tasas_glaukov ORDER BY id DESC LIMIT 1');
    let nextNum = 1;

    if (lastRow.rows.length > 0 && lastRow.rows[0].id_tasa) {
      const currentNum = parseInt(lastRow.rows[0].id_tasa.replace(/\D/g, ''), 10);
      if (!isNaN(currentNum)) nextNum = currentNum + 1;
    }

    const nuevoIdTasa = `T${String(nextNum).padStart(3, '0')}`;

    const objetoTasasCompleto = {
      ...tasas,
      CREATED_AT: tasas.CREATED_AT || new Date().toISOString()
    };

    const query = `
      INSERT INTO tasas_glaukov (id_tasa, tasas)
      VALUES ($1, $2)
      RETURNING *;
    `;
    const result = await pool.query(query, [nuevoIdTasa, JSON.stringify(objetoTasasCompleto)]);

    return res.json({
      success: true,
      id_tasa: nuevoIdTasa,
      registro: result.rows[0]
    });
  } catch (error) {
    console.error('[Error Guardar en tasas_glaukov]:', error.message);
    return res.status(500).json({ success: false, error: error.message });
  }
}

// 6. SOLO PUBLICAR
export async function publicarLoteSolo(req, res) {
  try {
    const { modo_prueba_global } = req.body;

    const lotes = await pool.query('SELECT * FROM tasas_glaukov ORDER BY id DESC LIMIT 2');
    if (lotes.rows.length === 0) {
      return res.status(400).json({ success: false, error: 'No hay lotes guardados en tasas_glaukov para publicar.' });
    }

    const loteActual = lotes.rows[0];
    const loteAnterior = lotes.rows[1] || lotes.rows[0];

    let sociosRes;
    try {
      sociosRes = await pool.query('SELECT * FROM perfiles_glaukov ORDER BY nombre ASC');
    } catch (e) {
      sociosRes = await pool.query('SELECT * FROM perfiles ORDER BY nombre ASC');
    }
    const socios = sociosRes.rows;

    const resultados = [];

    for (const socio of socios) {
      if (socio.activo === false) continue;

      try {
        const datosCalculadosRaw = calcularTasasCartelera(socio, loteActual, loteAnterior);
        const datosCalculados = normalizarParaRenderer(datosCalculadosRaw);
        
        const pngBuffer = await generarCarteleraPNG(datosCalculados);
        const base64Image = pngBuffer.toString('base64');

        const jidDestino = modo_prueba_global
          ? JID_GRUPO_PRUEBA
          : (socio.id_grupo || JID_GRUPO_PRUEBA);

        const caption = `${modo_prueba_global ? '🧪 [PRUEBA]' : '🚀 [OFICIAL]'} Cartelera *${datosCalculados.nombre_socio || socio.nombre}* — Lote: ${loteActual.id_tasa}`;

        await enviarMediaEvolution(jidDestino, base64Image, caption);
        resultados.push({ socio: socio.nombre, status: 'ok' });
      } catch (err) {
        resultados.push({ socio: socio.nombre, status: 'error', error: err.message });
      }
    }

    return res.json({
      success: true,
      id_tasa: loteActual.id_tasa,
      procesados: resultados.length,
      detalle: resultados
    });
  } catch (error) {
    console.error('[Error Publicar Lote]:', error.message);
    return res.status(500).json({ success: false, error: error.message });
  }
}

// 7. RENDERIZAR PNG
export async function renderCartelera(req, res) {
  try {
    const socioNombre = req.params.socio || req.body.socio;

    let socioRes;
    try {
      socioRes = await pool.query('SELECT * FROM perfiles_glaukov WHERE LOWER(nombre) = LOWER($1)', [socioNombre]);
    } catch (e) {
      socioRes = await pool.query('SELECT * FROM perfiles WHERE LOWER(nombre) = LOWER($1)', [socioNombre]);
    }
    if (socioRes.rows.length === 0) return res.status(404).send('Socio no encontrado');

    const lotes = await pool.query('SELECT * FROM tasas_glaukov ORDER BY id DESC LIMIT 2');
    const loteActual = lotes.rows[0] || {};
    const loteAnterior = lotes.rows[1] || loteActual;

    const datosCalculadosRaw = calcularTasasCartelera(socioRes.rows[0], loteActual, loteAnterior);
    const datosCalculados = normalizarParaRenderer(datosCalculadosRaw);
    
    const pngBuffer = await generarCarteleraPNG(datosCalculados);

    res.setHeader('Content-Type', 'image/png');
    return res.send(pngBuffer);
  } catch (error) {
    console.error('[Error renderCartelera]:', error.message);
    return res.status(500).send('Error generando PNG');
  }
}

// 8. DISPARAR SOCIO
export async function dispararCartelera(req, res) {
  try {
    const { socio, modo_prueba } = req.body;

    let socioRes;
    try {
      socioRes = await pool.query('SELECT * FROM perfiles_glaukov WHERE LOWER(nombre) = LOWER($1)', [socio]);
    } catch (e) {
      socioRes = await pool.query('SELECT * FROM perfiles WHERE LOWER(nombre) = LOWER($1)', [socio]);
    }

    if (socioRes.rows.length === 0) {
      return res.status(404).json({ success: false, error: 'Socio no encontrado' });
    }

    const socioData = socioRes.rows[0];
    const lotes = await pool.query('SELECT * FROM tasas_glaukov ORDER BY id DESC LIMIT 2');
    const loteActual = lotes.rows[0] || {};
    const loteAnterior = lotes.rows[1] || loteActual;

    const datosCalculadosRaw = calcularTasasCartelera(socioData, loteActual, loteAnterior);
    const datosCalculados = normalizarParaRenderer(datosCalculadosRaw);
    
    const pngBuffer = await generarCarteleraPNG(datosCalculados);
    const base64Image = pngBuffer.toString('base64');

    // MODO PRUEBA / GRUPO DESTINO
    const isTestMode = modo_prueba !== undefined ? modo_prueba : socioData.modo_prueba;
    const jidDestino = isTestMode
      ? JID_GRUPO_PRUEBA
      : (socioData.id_grupo || JID_GRUPO_PRUEBA);

    const caption = `${isTestMode ? '🧪 [PRUEBA]' : '🚀 [OFICIAL]'} Cartelera *${socioData.nombre}* — Lote: ${loteActual.id_tasa || 'T000'}`;

    const evoRes = await enviarMediaEvolution(jidDestino, base64Image, caption);

    return res.json({ success: true, socio: socioData.nombre, evolutionResponse: evoRes });
  } catch (error) {
    console.error('[Error dispararCartelera]:', error.message);
    return res.status(500).json({ success: false, error: error.message });
  }
}

// 9. DISPARAR LOTE COMPLETO
export async function dispararLoteGeneral(req, res) {
  return await publicarLoteSolo(req, res);
}

// 10. CALCULAR TASAS SOCIO
export async function calcularTasas(req, res) {
  try {
    const socioNombre = req.params.socio;
    let socioRes;
    try {
      socioRes = await pool.query('SELECT * FROM perfiles_glaukov WHERE LOWER(nombre) = LOWER($1)', [socioNombre]);
    } catch (e) {
      socioRes = await pool.query('SELECT * FROM perfiles WHERE LOWER(nombre) = LOWER($1)', [socioNombre]);
    }

    if (socioRes.rows.length === 0) {
      return res.status(404).json({ success: false, error: 'Socio no encontrado' });
    }

    const lotes = await pool.query('SELECT * FROM tasas_glaukov ORDER BY id DESC LIMIT 2');
    const loteActual = lotes.rows[0] || {};
    const loteAnterior = lotes.rows[1] || loteActual;

    const datosCalculadosRaw = calcularTasasCartelera(socioRes.rows[0], loteActual, loteAnterior);
    const datosCalculados = normalizarParaRenderer(datosCalculadosRaw);
    
    return res.json({ success: true, datos: datosCalculados });
  } catch (error) {
    console.error('[Error calcularTasas]:', error.message);
    return res.status(500).json({ success: false, error: error.message });
  }
}

// 11. ELIMINAR LOTE POR ID_TASA
export async function eliminarLotePorId(req, res) {
  try {
    const { id_tasa } = req.params;

    if (!id_tasa) {
      return res.status(400).json({ success: false, error: 'Debe proporcionar un ID de lote válido.' });
    }

    const result = await pool.query('DELETE FROM tasas_glaukov WHERE LOWER(id_tasa) = LOWER($1) RETURNING *', [id_tasa]);

    if (result.rowCount === 0) {
      return res.status(404).json({ success: false, error: `No se encontró el lote ${id_tasa}.` });
    }

    return res.json({
      success: true,
      message: `Lote ${id_tasa} eliminado correctamente.`
    });
  } catch (error) {
    console.error('[Error eliminarLotePorId]:', error.message);
    return res.status(500).json({ success: false, error: error.message });
  }
}
// 12. REENVIAR NOTIFICACIÓN DE TASAS REMESAS (Limpio y directo)
export async function reinterpretarNotificacionRemesas(req, res) {
  try {
    // 1. Obtener el último ID de tasa registrado
    const loteRes = await pool.query('SELECT id_tasa FROM tasas_glaukov ORDER BY id DESC LIMIT 1');
    const idTasa = loteRes.rows[0]?.id_tasa || 'T000';

    // 2. Insertar únicamente en las columnas reales de la tabla
    const query = `
      INSERT INTO notificaciones_tasas (id_tasa, creado_en)
      VALUES ($1, NOW())
      RETURNING *;
    `;
    const result = await pool.query(query, [idTasa]);

    console.log('⚡ [TRIGGER REMESAS ACTIVADO]:', result.rows[0]);
    return res.json({ success: true, notificacion: result.rows[0] });

  } catch (error) {
    console.error('❌ [Error reinterpretarNotificacionRemesas]:', error.message);
    return res.status(500).json({ success: false, error: error.message });
  }
}
