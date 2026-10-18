// Isla chart: gráficos en SVG (bar, pie o donut) sin dependencias.
export function mount(el, props) {
  var type = props.type || 'bar';
  var data = (props.data || '10,20,30,40').split(',').map(function (s) { return Number(s.trim()) || 0; });
  var labels = (props.labels || '').split(',').map(function (s) { return s.trim(); });
  var title = props.title || '';
  el.classList.add('chart');
  el.innerHTML = '';
  if (title) {
    var t = document.createElement('div');
    t.className = 'chart-title';
    t.textContent = title;
    el.appendChild(t);
  }
  if (type === 'pie' || type === 'donut') {
    renderPie(el, data, labels, type === 'donut');
  } else {
    renderBar(el, data, labels);
  }
}

var PALETTE = ['#2563eb', '#16a34a', '#db2777', '#f59e0b', '#7c3aed', '#0ea5e9', '#ef4444', '#14b8a6'];

function renderBar(el, data, labels) {
  var w = 340, h = 200, padL = 34, padR = 10, padT = 14, padB = 28;
  var max = Math.max.apply(null, data.concat([1]));
  var nice = niceMax(max);
  var plotW = w - padL - padR, plotH = h - padT - padB;
  var svg = svgEl(w, h, labels, data, false);
  // Gradiente único por gráfico (evita colisiones de id si hay varios).
  var gid = 'gbar-' + Math.random().toString(36).slice(2, 8);
  var NS = 'http://www.w3.org/2000/svg';
  var defs = document.createElementNS(NS, 'defs');
  var grad = document.createElementNS(NS, 'linearGradient');
  grad.setAttribute('id', gid);
  grad.setAttribute('x1', '0'); grad.setAttribute('y1', '0');
  grad.setAttribute('x2', '0'); grad.setAttribute('y2', '1');
  var s1 = document.createElementNS(NS, 'stop');
  s1.setAttribute('offset', '0%'); s1.setAttribute('stop-color', 'var(--accent)');
  var s2 = document.createElementNS(NS, 'stop');
  s2.setAttribute('offset', '100%'); s2.setAttribute('stop-color', 'var(--accent)'); s2.setAttribute('stop-opacity', '0.5');
  grad.appendChild(s1); grad.appendChild(s2);
  defs.appendChild(grad);
  svg.appendChild(defs);

  // Líneas guía + etiquetas de eje Y.
  var steps = 4;
  for (var s = 0; s <= steps; s++) {
    var v = (nice / steps) * s;
    var y = padT + plotH - (v / nice) * plotH;
    var line = document.createElementNS('http://www.w3.org/2000/svg', 'line');
    line.setAttribute('x1', padL); line.setAttribute('y1', y);
    line.setAttribute('x2', w - padR); line.setAttribute('y2', y);
    line.setAttribute('stroke', 'var(--border)');
    line.setAttribute('stroke-width', '1');
    if (s > 0) line.setAttribute('stroke-dasharray', '3 3');
    svg.appendChild(line);
    var yl = document.createElementNS('http://www.w3.org/2000/svg', 'text');
    yl.setAttribute('x', padL - 6); yl.setAttribute('y', y + 3);
    yl.setAttribute('text-anchor', 'end'); yl.setAttribute('font-size', '9');
    yl.setAttribute('fill', 'var(--muted)');
    yl.textContent = String(Math.round(v));
    svg.appendChild(yl);
  }

  var slot = plotW / data.length;
  var bw = Math.min(slot * 0.6, 46);
  data.forEach(function (val, i) {
    var bh = (val / nice) * plotH;
    var x = padL + slot * i + (slot - bw) / 2;
    var y = padT + plotH - bh;
    var rect = document.createElementNS('http://www.w3.org/2000/svg', 'rect');
    rect.setAttribute('x', x); rect.setAttribute('y', y);
    rect.setAttribute('width', bw); rect.setAttribute('height', Math.max(bh, 0));
    rect.setAttribute('rx', '4'); rect.setAttribute('fill', 'url(#' + gid + ')');
    svg.appendChild(rect);
    // Valor sobre la barra.
    var vl = document.createElementNS('http://www.w3.org/2000/svg', 'text');
    vl.setAttribute('x', x + bw / 2); vl.setAttribute('y', y - 4);
    vl.setAttribute('text-anchor', 'middle'); vl.setAttribute('font-size', '10');
    vl.setAttribute('font-weight', '600'); vl.setAttribute('fill', 'var(--fg)');
    vl.textContent = String(val);
    svg.appendChild(vl);
    // Etiqueta X.
    if (labels[i]) {
      var xl = document.createElementNS('http://www.w3.org/2000/svg', 'text');
      xl.setAttribute('x', x + bw / 2); xl.setAttribute('y', h - padB + 16);
      xl.setAttribute('text-anchor', 'middle'); xl.setAttribute('font-size', '10');
      xl.setAttribute('fill', 'var(--muted)');
      xl.textContent = labels[i];
      svg.appendChild(xl);
    }
  });
  el.appendChild(svg);
}

function renderPie(el, data, labels, donut) {
  var size = 220;
  var cx = size / 2, cy = size / 2;
  var r = size / 2 - 12;
  var inner = donut ? r * 0.58 : 0;
  var total = data.reduce(function (a, b) { return a + b; }, 0) || 1;
  var svg = svgEl(size, size, labels, data, true);
  var angle = -Math.PI / 2;
  data.forEach(function (v, i) {
    var slice = (v / total) * 2 * Math.PI;
    var color = PALETTE[i % PALETTE.length];
    var path = document.createElementNS('http://www.w3.org/2000/svg', 'path');
    var x1 = cx + r * Math.cos(angle), y1 = cy + r * Math.sin(angle);
    var x2 = cx + r * Math.cos(angle + slice), y2 = cy + r * Math.sin(angle + slice);
    var large = slice > Math.PI ? 1 : 0;
    var d;
    if (donut) {
      var ix1 = cx + inner * Math.cos(angle + slice), iy1 = cy + inner * Math.sin(angle + slice);
      var ix2 = cx + inner * Math.cos(angle), iy2 = cy + inner * Math.sin(angle);
      d = 'M ' + x1 + ' ' + y1 + ' A ' + r + ' ' + r + ' 0 ' + large + ' 1 ' + x2 + ' ' + y2 +
        ' L ' + ix1 + ' ' + iy1 + ' A ' + inner + ' ' + inner + ' 0 ' + large + ' 0 ' + ix2 + ' ' + iy2 + ' Z';
    } else {
      d = 'M ' + cx + ' ' + cy + ' L ' + x1 + ' ' + y1 + ' A ' + r + ' ' + r + ' 0 ' + large + ' 1 ' + x2 + ' ' + y2 + ' Z';
    }
    path.setAttribute('d', d);
    path.setAttribute('fill', color);
    path.setAttribute('stroke', 'var(--bg)');
    path.setAttribute('stroke-width', '2');
    svg.appendChild(path);

    // Porcentaje dentro del sector (si es suficientemente grande).
    var pct = Math.round((v / total) * 100);
    if (pct >= 6) {
      var mid = angle + slice / 2;
      var lr = donut ? (r + inner) / 2 : r * 0.62;
      var tx = cx + lr * Math.cos(mid), ty = cy + lr * Math.sin(mid);
      var tl = document.createElementNS('http://www.w3.org/2000/svg', 'text');
      tl.setAttribute('x', tx); tl.setAttribute('y', ty + 3);
      tl.setAttribute('text-anchor', 'middle'); tl.setAttribute('font-size', '10');
      tl.setAttribute('font-weight', '700'); tl.setAttribute('fill', '#fff');
      tl.textContent = pct + '%';
      svg.appendChild(tl);
    }
    angle += slice;
  });
  // Total en el centro (donut).
  if (donut) {
    var ct = document.createElementNS('http://www.w3.org/2000/svg', 'text');
    ct.setAttribute('x', cx); ct.setAttribute('y', cy + 4);
    ct.setAttribute('text-anchor', 'middle'); ct.setAttribute('font-size', '18');
    ct.setAttribute('font-weight', '700'); ct.setAttribute('fill', 'var(--fg)');
    ct.textContent = String(total);
    svg.appendChild(ct);
  }
  el.appendChild(svg);

  // Leyenda.
  if (labels.length) {
    var leg = document.createElement('div');
    leg.className = 'chart-legend';
    labels.forEach(function (l, i) {
      var item = document.createElement('span');
      item.className = 'chart-legend-item';
      var dot = document.createElement('span');
      dot.className = 'chart-legend-dot';
      dot.style.background = PALETTE[i % PALETTE.length];
      item.appendChild(dot);
      item.appendChild(document.createTextNode((l || '') + ' (' + data[i] + ')'));
      leg.appendChild(item);
    });
    el.appendChild(leg);
  }
}

// svgEl crea el <svg> accesible con la descripción textual.
function svgEl(w, h, labels, data, ro) {
  var svg = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
  svg.setAttribute('width', '100%');
  svg.setAttribute('viewBox', '0 0 ' + w + ' ' + h);
  svg.setAttribute('role', 'img');
  var desc = data.map(function (v, i) { return (labels[i] ? labels[i] + ': ' : '') + v; }).join(', ');
  svg.setAttribute('aria-label', desc);
  // <title> da nombre accesible al SVG (además de aria-label).
  var ttl = document.createElementNS('http://www.w3.org/2000/svg', 'title');
  ttl.textContent = desc;
  svg.appendChild(ttl);
  return svg;
}

// niceMax redondea el máximo a un valor "bonito" para el eje.
function niceMax(max) {
  if (max <= 0) return 1;
  var pow = Math.pow(10, Math.floor(Math.log(max) / Math.LN10));
  var n = Math.ceil(max / pow);
  if (n <= 1) n = 1; else if (n <= 2) n = 2; else if (n <= 5) n = 5; else n = 10;
  return n * pow;
}
