export function aplicarReglaPrecisionTasa(val) {
  if (val === null || val === undefined || isNaN(val) || val === 0) return 0;
  const num = parseFloat(val);
  if (num === 0) return 0;

  const signo = num < 0 ? -1 : 1;
  const vRound = Math.round(Math.abs(num) * 1e8) / 1e8;

  let res = 0;
  if (vRound > 99.99) {
    res = Math.trunc(vRound);
  } else if (vRound >= 10.0) {
    res = Math.trunc((vRound + 0.0000001) * 100) / 100;
  } else {
    const magnitud = Math.floor(Math.log10(vRound));
    const factor = Math.pow(10, 2 - magnitud);
    res = Math.trunc((vRound + 0.0000001) * factor) / factor;
  }

  return signo * res;
}

export function normalizarNumero(valor) {
  if (valor === null || valor === undefined) return 0;
  const num = parseFloat(String(valor).replace(/,/g, ''));
  return isNaN(num) ? 0 : num;
}

export function truncarTasaOficial(val) {
  const num = aplicarReglaPrecisionTasa(val);
  return num === 0 ? "0" : num.toString();
}
