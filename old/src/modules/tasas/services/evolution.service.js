import { env } from '../../../config/env.js';

/**
 * Envía la cartelera PNG en Base64 al grupo de WhatsApp correspondiente vía Evolution API
 * Instancia: GLAUK
 */
export async function enviarCarteleraWhatsApp(jidDestino, base64Image, caption = '') {
  if (!jidDestino) {
    throw new Error('No se proporcionó un JID de WhatsApp válido.');
  }

  // Nombre exacto de tu instancia en Evolution API
  const instanceName = env.EVOLUTION_INSTANCE || 'GLAUK';
  const url = `${env.EVOLUTION_API_URL}/message/sendMedia/${instanceName}`;

  const payload = {
    number: jidDestino,
    mediaMessage: {
      mediatype: 'image',
      caption: caption,
      media: `data:image/png;base64,${base64Image}`
    }
  };

  const response = await fetch(url, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'apikey': env.AUTHENTICATION_API_KEY
    },
    body: JSON.stringify(payload)
  });

  const data = await response.json();

  if (!response.ok) {
    console.error('[Evolution API Error]:', data);
    throw new Error(`Error en Evolution API: ${data.message || JSON.stringify(data)}`);
  }

  return data;
}
