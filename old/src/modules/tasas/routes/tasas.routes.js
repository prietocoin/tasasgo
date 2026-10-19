import express from 'express';
import {
  calcularTasas,
  renderCartelera,
  dispararCartelera,
  dispararLoteGeneral,
  obtenerTasasMonitor,
  obtenerListaSocios,
  actualizarMonedasSocio,
  traerBorradorHoo,
  guardarLoteSolo,
  publicarLoteSolo,
  eliminarLotePorId,
  reinterpretarNotificacionRemesas
} from '../controllers/tasas.controller.js';

const router = express.Router();

// Endpoints Dashboard
router.get('/monitor', obtenerTasasMonitor);
router.get('/socios', obtenerListaSocios);
router.get('/hoo-borrador', traerBorradorHoo);

router.put('/socios/:socio', actualizarMonedasSocio);
router.post('/guardar-lote', guardarLoteSolo);
router.post('/publicar-lote', publicarLoteSolo);

// 🚀 Endpoint Notificar Remesas
router.post('/notificar-remesas', reinterpretarNotificacionRemesas);

// Endpoints Producción Existentes
router.get('/calcular/:socio', calcularTasas);
router.get('/render/:socio', renderCartelera);
router.post('/render', renderCartelera);
router.post('/disparar', dispararCartelera);
router.post('/disparar-lote', dispararLoteGeneral);
router.delete('/lote/:id_tasa', eliminarLotePorId);

export default router;
