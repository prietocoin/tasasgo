import satori from 'satori';
import { Resvg } from '@resvg/resvg-js';
import fs from 'fs';
import path from 'path';
import { renderCarteleraTemplate } from '../templates/cartelera.template.js';

let fontBuffer = null;
const assetsCache = {
  logo: null,
  background: null,
  flags: {},
};

const MAPA_BANDERAS_ISO = {
  ARS: 'ar', BRL: 'br', CLP: 'cl', COP: 'co', MXN: 'mx',
  PEN: 'pe', PYG: 'py', VES: 've', USD: 'us', USDT: 'us',
  BOB: 'bo', EUR: 'eu', ECU: 'ec', PAN: 'pa',
};

function obtenerFuente() {
  if (!fontBuffer) {
    const fontPath = path.join(process.cwd(), 'assets/fonts/Inter-Bold.ttf');
    if (fs.existsSync(fontPath)) {
      fontBuffer = fs.readFileSync(fontPath);
    } else {
      console.error('❌ CRÍTICO: No se encontró la fuente en:', fontPath);
    }
  }
  return fontBuffer;
}

function fileToBase64(filePath) {
  if (!fs.existsSync(filePath)) return null;
  const buffer = fs.readFileSync(filePath);
  const ext = path.extname(filePath).toLowerCase().replace('.', '');
  const mimeType = ext === 'svg' ? 'image/svg+xml' : `image/${ext === 'png' ? 'png' : 'jpeg'}`;
  return `data:${mimeType};base64,${buffer.toString('base64')}`;
}

export function obtenerAssetBase64(nombre) {
  if (!assetsCache[nombre]) {
    const filePng = path.join(process.cwd(), `assets/${nombre}.png`);
    const fileSvg = path.join(process.cwd(), `assets/${nombre}.svg`);
    const fileJpg = path.join(process.cwd(), `assets/${nombre}.jpg`);
    const fileJpeg = path.join(process.cwd(), `assets/${nombre}.jpeg`);

    assetsCache[nombre] = fileToBase64(filePng) || fileToBase64(fileSvg) || fileToBase64(fileJpg) || fileToBase64(fileJpeg);
  }
  return assetsCache[nombre];
}

export function obtenerBanderaBase64(code) {
  if (!code) return null;
  const codeUpper = code.toUpperCase();
  const filename = MAPA_BANDERAS_ISO[codeUpper] || code.toLowerCase().substring(0, 2);

  if (!assetsCache.flags[filename]) {
    const flagPng = path.join(process.cwd(), `assets/flags/${filename}.png`);
    const flagSvg = path.join(process.cwd(), `assets/flags/${filename}.svg`);
    assetsCache.flags[filename] = fileToBase64(flagPng) || fileToBase64(flagSvg);
  }
  return assetsCache.flags[filename];
}

export async function generarCarteleraPNG(datosCalculados) {
  const fontData = obtenerFuente();

  if (!fontData) {
    throw new Error('No se pudo cargar la fuente Inter-Bold.ttf necesaria para Satori.');
  }

  const totalTarjetas = datosCalculados.tarjetas_paises?.length || 0;
  const filas = Math.ceil(totalTarjetas / 2);

  // CÁLCULO DE ALTURA DINÁMICA EXACTA (ENTERO REDONDEADO)
  // Header (110px) + Footer (90px) + Margen/Padding = 260px base
  // Cada fila de tarjetas (150px alto + 20px gap) = 170px
  const calculatedHeight = Math.round(260 + (filas * 170));
  const finalHeight = Math.max(450, calculatedHeight);

  const svg = await satori(renderCarteleraTemplate(datosCalculados), {
    width: 1080,
    height: finalHeight,
    fonts: [
      {
        name: 'Inter',
        data: fontData,
        weight: 700,
        style: 'normal',
      },
    ],
  });

  const resvg = new Resvg(svg, {
    fitTo: {
      mode: 'width',
      value: 1080,
    },
  });

  return resvg.render().asPng();
}
