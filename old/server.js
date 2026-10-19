import express from 'express';
import cors from 'cors';
import path from 'path';
import { env } from './src/config/env.js';
import { apiProxy } from './src/proxy/api-proxy.js';

const app = express();

app.use(cors());

// 🟢 PERMITIR IFRAME DESDE LAB.JAIROKOV.COM
// Va antes que todo lo demás para que la cabecera viaje también en las respuestas
// que atraviesan el proxy.
app.use((req, res, next) => {
  res.setHeader(
    'Content-Security-Policy',
    "frame-ancestors 'self' https://lab.jairokov.com"
  );
  res.removeHeader('X-Frame-Options');
  next();
});

// 🔀 TODO lo que empieza por /api pertenece al backend Go (TasasGo).
// Se monta antes de express.json para reenviar el cuerpo tal como llegó, y antes
// de los estáticos para que ningún archivo del panel pueda tapar una ruta de la
// API. Este servidor ya no abre ninguna conexión a PostgreSQL.
app.use('/api', apiProxy());

app.use(express.json({ limit: '10mb' }));

// Rutas estáticas (ahora saldrán con las cabeceras de iframe aplicadas)
app.use(express.static('public'));
app.use('/flags', express.static(path.join(process.cwd(), 'assets/flags')));

// Health Check
app.get('/health', (req, res) => {
  res.json({ status: 'OK', service: 'TasasHub Engine ⚡', timestamp: new Date() });
});

app.listen(env.PORT, '0.0.0.0', () => {
  console.log(`[TasasHub ⚡] Servidor ejecutándose en el puerto ${env.PORT}`);
  console.log(`[TasasHub 🔀] /api se reenvía a ${env.API_URL}`);
});
