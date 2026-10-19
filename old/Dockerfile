# 1. Etapa de compilación
FROM node:20-alpine AS builder
WORKDIR /app

COPY package*.json ./
RUN npm install

COPY . .

# 2. Etapa de ejecución ligera
FROM node:20-alpine AS runner
WORKDIR /app

ENV NODE_ENV=production
ENV PORT=3002

COPY --from=builder /app ./

EXPOSE 3002
CMD ["node", "server.js"]
