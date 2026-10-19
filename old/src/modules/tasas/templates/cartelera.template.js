import { obtenerAssetBase64, obtenerBanderaBase64 } from '../services/renderer.service.js';

const h = (type, props = {}, ...children) => {
  const flatChildren = children.flat(Infinity).filter((c) => c !== null && c !== undefined && c !== false);
  return {
    type,
    props: {
      ...props,
      style: {
        ...(type === 'div' ? { display: 'flex' } : {}),
        ...(props.style || {}),
      },
      children: flatChildren.length === 1 ? flatChildren[0] : flatChildren,
    },
  };
};

function renderTrendBadge(trend) {
  if (trend === 'up') {
    return h(
      'div',
      {
        style: {
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          backgroundColor: 'rgba(52, 211, 153, 0.15)',
          border: '1px solid rgba(52, 211, 153, 0.35)',
          borderRadius: '8px',
          padding: '4px 8px',
        },
      },
      h(
        'svg',
        {
          width: 18,
          height: 18,
          viewBox: '0 0 24 24',
          fill: 'none',
          stroke: '#34d399',
          strokeWidth: 3,
          strokeLinecap: 'round',
          strokeLinejoin: 'round',
        },
        h('polyline', { points: '23 6 13.5 15.5 8.5 10.5 1 18' }),
        h('polyline', { points: '17 6 23 6 23 12' })
      )
    );
  }

  if (trend === 'down') {
    return h(
      'div',
      {
        style: {
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          backgroundColor: 'rgba(251, 113, 133, 0.15)',
          border: '1px solid rgba(251, 113, 133, 0.35)',
          borderRadius: '8px',
          padding: '4px 8px',
        },
      },
      h(
        'svg',
        {
          width: 18,
          height: 18,
          viewBox: '0 0 24 24',
          fill: 'none',
          stroke: '#fb7185',
          strokeWidth: 3,
          strokeLinecap: 'round',
          strokeLinejoin: 'round',
        },
        h('polyline', { points: '23 18 13.5 8.5 8.5 13.5 1 6' }),
        h('polyline', { points: '17 18 23 18 23 12' })
      )
    );
  }

  return h(
    'div',
    {
      style: {
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        backgroundColor: 'rgba(251, 191, 36, 0.15)',
        border: '1px solid rgba(251, 191, 36, 0.35)',
        borderRadius: '8px',
        padding: '4px 8px',
      },
    },
    h(
      'svg',
      {
        width: 18,
        height: 18,
        viewBox: '0 0 24 24',
        fill: 'none',
        stroke: '#fbbf24',
        strokeWidth: 3,
        strokeLinecap: 'round',
        strokeLinejoin: 'round',
      },
      h('line', { x1: '4', y1: '12', x2: '20', y2: '12' }),
      h('polyline', { points: '14 6 20 12 14 18' })
    )
  );
}

export function renderCarteleraTemplate(data) {
  const { nombre_socio, tarjetas_paises = [], hora_actualizacion, fecha_actualizacion, lote_tasa } = data;

  const bgImageSrc = obtenerAssetBase64('background') || obtenerAssetBase64('bg-template');
  const logoSrc = obtenerAssetBase64('logo');

  return h(
    'div',
    {
      style: {
        width: '100%',
        height: '100%',
        display: 'flex',
        flexDirection: 'column',
        backgroundColor: '#040a17',
        backgroundImage: bgImageSrc ? `url(${bgImageSrc})` : 'none',
        backgroundSize: 'cover',
        backgroundPosition: 'center top',
        backgroundRepeat: 'no-repeat',
        padding: '40px',
        color: '#ffffff',
        fontFamily: 'Inter',
        boxSizing: 'border-box',
      },
    },
    // Header
    h(
      'div',
      {
        style: {
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          marginBottom: '25px',
        },
      },
      h(
        'div',
        { style: { display: 'flex', alignItems: 'center', gap: '15px' } },
        logoSrc
          ? h('img', {
              src: logoSrc,
              height: 65,
              style: { objectFit: 'contain' },
            })
          : null,
        h(
          'div',
          { style: { display: 'flex', alignItems: 'center', fontSize: '36px', fontWeight: 'bold' } },
          h('span', { style: { color: '#ffffff', letterSpacing: '1px' } }, 'FUNDABLOCK'),
          h('span', { style: { color: '#d97706', margin: '0 12px' } }, '|'),
          h('span', { style: { color: '#94a3b8', fontWeight: 'normal' } }, nombre_socio)
        )
      ),
      h(
        'div',
        {
          style: {
            display: 'flex',
            fontSize: '14px',
            fontWeight: 'bold',
            color: '#fbbf24',
            backgroundColor: 'rgba(217, 119, 6, 0.15)',
            border: '1px solid rgba(217, 119, 6, 0.4)',
            padding: '6px 16px',
            borderRadius: '20px',
            letterSpacing: '1px',
          },
        },
        'OFICIAL'
      )
    ),
    // Grid de Países (Satori Compatible: flexGrow 0, flexShrink 0)
    h(
      'div',
      {
        style: {
          display: 'flex',
          flexDirection: 'row',
          flexWrap: 'wrap',
          gap: '20px',
          justifyContent: 'space-between',
          alignContent: 'flex-start',
          flexGrow: 0,
          flexShrink: 0,
          marginBottom: '20px',
        },
      },
      tarjetas_paises.map((item) => {
        const flagSrc = obtenerBanderaBase64(item.code);

        return h(
          'div',
          {
            key: item.code,
            style: {
              width: '48%',
              height: '150px',
              backgroundColor: 'rgba(11, 19, 38, 0.85)',
              border: '1.5px solid rgba(59, 130, 246, 0.35)',
              borderTop: '4px solid #3b82f6',
              borderRadius: '16px',
              padding: '14px 20px',
              display: 'flex',
              flexDirection: 'column',
              justifyContent: 'space-between',
              boxSizing: 'border-box',
            },
          },
          // Header Tarjeta
          h(
            'div',
            { style: { display: 'flex', alignItems: 'center', justifyContent: 'space-between' } },
            h(
              'div',
              { style: { display: 'flex', alignItems: 'center', gap: '12px' } },
              flagSrc
                ? h('img', {
                    src: flagSrc,
                    width: 36,
                    height: 26,
                    style: { borderRadius: '4px', objectFit: 'cover', border: '1px solid rgba(255,255,255,0.2)' },
                  })
                : null,
              h('span', { style: { fontSize: '24px', fontWeight: 'bold', color: '#ffffff' } }, item.nombre_pais)
            ),
            renderTrendBadge(item.trend)
          ),
          // Tasas Comprar / Vender
          h(
            'div',
            {
              style: {
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
                borderTop: '1px solid rgba(255, 255, 255, 0.1)',
                paddingTop: '8px',
              },
            },
            h(
              'div',
              { style: { display: 'flex', flexDirection: 'column' } },
              h('span', { style: { fontSize: '12px', color: '#94a3b8', textTransform: 'uppercase' } }, 'Comprar'),
              h('span', { style: { fontSize: '28px', fontWeight: 'bold', color: '#34d399' } }, String(item.compra))
            ),
            h('div', { style: { display: 'flex', width: '1px', height: '35px', backgroundColor: 'rgba(255, 255, 255, 0.15)' } }),
            h(
              'div',
              { style: { display: 'flex', flexDirection: 'column', alignItems: 'flex-end' } },
              h('span', { style: { fontSize: '12px', color: '#94a3b8', textTransform: 'uppercase' } }, 'Vender'),
              h('span', { style: { fontSize: '28px', fontWeight: 'bold', color: '#fb7185' } }, String(item.venta))
            )
          )
        );
      })
    ),
    // Footer Adherido
    h(
      'div',
      {
        style: {
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          marginTop: 'auto',
          paddingTop: '10px',
        },
      },
      h(
        'div',
        {
          style: {
            display: 'flex',
            alignItems: 'center',
            gap: '8px',
            backgroundColor: 'rgba(15, 23, 42, 0.85)',
            border: '1px solid rgba(255, 255, 255, 0.15)',
            padding: '6px 16px',
            borderRadius: '20px',
            color: '#94a3b8',
            fontSize: '15px',
          },
        },
        h('span', { style: { color: '#ffffff', fontWeight: 'bold' } }, 'TasasHub Engine'),
        h('span', { style: { color: '#3b82f6' } }, '•'),
        h('span', {}, 'Fundablock')
      ),
      h(
        'div',
        {
          style: {
            display: 'flex',
            alignItems: 'center',
            gap: '10px',
            backgroundColor: 'rgba(15, 23, 42, 0.9)',
            border: '1px solid rgba(59, 130, 246, 0.4)',
            padding: '6px 18px',
            borderRadius: '20px',
            color: '#ffffff',
            fontSize: '15px',
          },
        },
        h('span', { style: { color: '#fbbf24', fontWeight: 'bold' } }, `Lote: ${lote_tasa || 'T000'}`),
        h('span', { style: { color: '#3b82f6' } }, '|'),
        h('span', { style: { color: '#e2e8f0' } }, `${fecha_actualizacion || ''} • ${hora_actualizacion}`)
      )
    )
  );
}
