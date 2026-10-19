import pool from '../../../config/db.js';

export async function obtenerUltimoLoteTasas() {
  const query = `
    SELECT id_tasa, tasas, created_at
    FROM tasas_glaukov
    ORDER BY created_at DESC
    LIMIT 1;
  `;
  const { rows } = await pool.query(query);
  if (rows.length === 0) {
    throw new Error('No se encontraron lotes de tasas en la tabla tasas_glaukov.');
  }
  return rows[0];
}

export async function obtenerUltimosDosLotesTasas() {
  const query = `
    SELECT id_tasa, tasas, created_at
    FROM tasas_glaukov
    ORDER BY created_at DESC
    LIMIT 2;
  `;
  const { rows } = await pool.query(query);
  if (rows.length === 0) {
    throw new Error('No se encontraron lotes de tasas en la tabla tasas_glaukov.');
  }
  return {
    loteActual: rows[0],
    loteAnterior: rows[1] || null,
  };
}

export async function obtenerPerfilSocio(nombreSocio) {
  if (!nombreSocio) {
    throw new Error('El nombre del socio no fue proporcionado en la petición.');
  }

  const query = `
    SELECT *
    FROM perfiles_glaukov
    WHERE LOWER(nombre) = LOWER($1)
    LIMIT 1;
  `;
  const { rows } = await pool.query(query, [String(nombreSocio).trim()]);
  if (rows.length === 0) {
    throw new Error(`Socio "${nombreSocio}" no encontrado en perfiles_glaukov.`);
  }
  return rows[0];
}
export async function obtenerTodosLosSocios() {
  const query = `SELECT nombre, moneda_base, monedas FROM perfiles_glaukov ORDER BY nombre ASC;`;
  const { rows } = await pool.query(query);
  return rows;
}

export async function actualizarConfiguracionSocioBD(nombreSocio, monedas) {
  const query = `
    UPDATE perfiles_glaukov 
    SET monedas = $1 
    WHERE LOWER(nombre) = LOWER($2) 
    RETURNING nombre, monedas;
  `;
  const { rows } = await pool.query(query, [JSON.stringify(monedas), String(nombreSocio).trim()]);
  return rows[0];
}
