import dotenv from 'dotenv';
dotenv.config();

export const env = {
  PORT: process.env.PORT || 3002,
  // URL del backend Go. En Easypanel se define con la URL pública o interna de
  // tasasgo; en local se asume el puerto 3003 de siempre.
  API_URL: process.env.API_URL || 'http://localhost:3003',
  DB_HOST: process.env.DB_HOST || 'localhost',
  DB_PORT: Number(process.env.DB_PORT) || 5432,
  DB_USER: process.env.DB_USER || 'postgres',
  DB_PASSWORD: process.env.DB_PASSWORD || '',
  DB_NAME: process.env.DB_NAME || 'automat',
};
