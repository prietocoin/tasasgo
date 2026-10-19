import express from 'express';
import cors from 'cors';
import path from 'path';
import { env } from './src/config/env.js';
import tasasRoutes from './src/modules/tasas/routes/tasas.routes.js';

const app = express();

app.use(cors());
app.use(express.json({ limit: '10mb' }));

// 🟢 PERMITIR IFRAME DESDE LAB.JAIROKOV.COM
app.use((req, res, next) => {
  res.setHeader(
    'Content-Security-Policy',
    "frame-ancestors 'self' https://lab.jairokov.com"
  );
  res.removeHeader('X-Frame-Options');
  next();
});

// Rutas estáticas (ahora saldrán con las cabeceras de iframe aplicadas)
app.use(express.static('public'));
app.use('/flags', express.static(path.join(process.cwd(), 'assets/flags')));

// Health Check
app.get('/health', (req, res) => {
  res.json({ status: 'OK', service: 'TasasHub Engine ⚡', timestamp: new Date() });
});

// Enrutador principal de TasasHub
app.use('/api/v1/tasas', tasasRoutes);

app.listen(env.PORT, '0.0.0.0', () => {
  console.log(`[TasasHub ⚡] Servidor ejecutándose en el puerto ${env.PORT}`);
});
