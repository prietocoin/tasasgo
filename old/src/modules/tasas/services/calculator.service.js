import { normalizarNumero, aplicarReglaPrecisionTasa } from '../../../utils/formatters.js';

const BANDERAS_MAP = {
  COP: 'co', PEN: 'pe', ARS: 'ar', BRL: 'br', MXN: 'mx', EUR: 'eu',
  USD: 'us', USDT: 'us', PYUSD: 'us', PYG: 'py', VES: 've', CLP: 'cl', DOP: 'do',
  CRC: 'cr', BOB: 'bo', ECU: 'ec', PAN: 'pa', CAD: 'ca'
};

const NOMBRES_PAIS = {
  COP: 'Colombia', PEN: 'Peru', ARS: 'Argentina', BRL: 'Brazil',
  MXN: 'Mexico', EUR: 'Europa', USD: 'EEUU-Zelle', USDT: 'Tether',
  PYUSD: 'PayPal USD', PYG: 'Paraguay', VES: 'Venezuela', CLP: 'Chile', DOP: 'Dominicana',
  CRC: 'Costa Rica', BOB: 'Bolivia', ECU: 'Ecuador', PAN: 'Panama', CAD: 'Canada'
};

/**
 * Auxiliar para obtener la tasa base de la moneda nativa del socio (ej: PEN)
 */
function obtenerTasaMonedaBase(mapaTasas, monedaBase) {
  const code = (monedaBase || 'USDT').toUpperCase().trim();
  if (['USD', 'USDT', 'PYUSD', 'ECU', 'PAN'].includes(code)) return 1.0;

  const rawVal = mapaTasas[code];
  const numVal = normalizarNumero(rawVal);
  return numVal > 0 ? numVal : 1.0;
}

export function calcularTasasCartelera(perfil, loteActual, loteAnterior = null) {
  const { nombre, id_grupo, moneda_base, monedas = {} } = perfil;
  const mapaTasasActuales = loteActual.tasas || {};
  const mapaTasasAnteriores = loteAnterior?.tasas || {};

  // 1. TASA BASE DE LA MONEDA DEL SOCIO
  const baseSocioActual = obtenerTasaMonedaBase(mapaTasasActuales, moneda_base);
  const baseSocioAnterior = obtenerTasaMonedaBase(mapaTasasAnteriores, moneda_base);

  const tarjetasPaises = [];

  for (const [code, config] of Object.entries(monedas)) {
    if (!config || config.activo === false) continue;

    const codeUpper = code.toUpperCase();
    const tasaBaseRaw = mapaTasasActuales[codeUpper];

    if (tasaBaseRaw === undefined && !['USD', 'USDT', 'PYUSD', 'ECU', 'PAN'].includes(codeUpper)) continue;

    const tasaBasePais = normalizarNumero(tasaBaseRaw || 1.0);
    if (tasaBasePais === 0) continue;

    // 2. TRIANGULACIÓN REAL
    const crossBaseActual = tasaBasePais / baseSocioActual;

    // 3. TENDENCIA TRIANGULADA
    const tasaAnteriorRaw = mapaTasasAnteriores[codeUpper];
    let trend = 'equal';

    if (tasaAnteriorRaw !== undefined) {
      const tasaBasePaisAnt = normalizarNumero(tasaAnteriorRaw);
      const crossBaseAnterior = tasaBasePaisAnt / baseSocioAnterior;

      if (crossBaseActual > crossBaseAnterior) trend = 'up';
      else if (crossBaseActual < crossBaseAnterior) trend = 'down';
    }

    // 4. APLICACIÓN DE PORCENTAJES SOBRE TASA BASE
    const pctDeposito = Math.abs(normalizarNumero(config.porcentaje?.deposito || 0));
    const pctPago = Math.abs(normalizarNumero(config.porcentaje?.pago || 0));

    const factorD = 1 + (pctDeposito / 100);
    const factorP = 1 + (pctPago / 100);

    const rawCompra = crossBaseActual * factorD;
    const rawVenta = crossBaseActual / factorP;

    tarjetasPaises.push({
      code: codeUpper,
      nombre_pais: NOMBRES_PAIS[codeUpper] || codeUpper,
      bandera: BANDERAS_MAP[codeUpper] || 'us',
      compra: aplicarReglaPrecisionTasa(rawCompra),
      venta: aplicarReglaPrecisionTasa(rawVenta),
      trend,
    });
  }

  const ahoraVE = new Date();
  const fechaCorta = ahoraVE.toLocaleDateString('es-VE', {
    timeZone: 'America/Caracas',
    day: '2-digit',
    month: 'short'
  }).replace('.', '');

  const horaVE = ahoraVE.toLocaleTimeString('es-VE', {
    timeZone: 'America/Caracas',
    hour: '2-digit',
    minute: '2-digit',
    hour12: false
  });

  return {
    nombre_socio: nombre,
    id_grupo,
    moneda_base: moneda_base || 'USDT',
    lote_tasa: loteActual.id_tasa || 'T000',
    fecha_actualizacion: fechaCorta,
    hora_actualizacion: horaVE,
    tarjetas_paises: tarjetasPaises
  };
}
