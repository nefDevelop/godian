// Isla qr-code: genera un código QR en SVG (sin dependencias).
// Implementación compacta: modo byte, EC nivel L, versión 1 (21x21), máscara 0.
export function mount(el, props) {
  var text = props.text || (el.textContent || '').trim() || 'https://godian';
  var size = Number(props.size || 160);
  el.classList.add('qr-code');
  el.innerHTML = '';
  try {
    var matrix = buildQR(text);
    var n = matrix.length;
    var cell = size / (n + 8);
    var off = 4 * cell;
    var svg = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
    svg.setAttribute('width', size);
    svg.setAttribute('height', size);
    svg.setAttribute('viewBox', '0 0 ' + size + ' ' + size);
    svg.setAttribute('role', 'img');
    svg.setAttribute('aria-label', 'Código QR: ' + text);
    var ttl = document.createElementNS('http://www.w3.org/2000/svg', 'title');
    ttl.textContent = 'Código QR: ' + text;
    svg.appendChild(ttl);
    var g = document.createElementNS('http://www.w3.org/2000/svg', 'g');
    for (var y = 0; y < n; y++) {
      for (var x = 0; x < n; x++) {
        if (matrix[y][x]) {
          var r = document.createElementNS('http://www.w3.org/2000/svg', 'rect');
          r.setAttribute('x', off + x * cell);
          r.setAttribute('y', off + y * cell);
          r.setAttribute('width', cell);
          r.setAttribute('height', cell);
          r.setAttribute('rx', cell * 0.18);
          r.setAttribute('fill', 'var(--qr-fg, var(--accent))');
          g.appendChild(r);
        }
      }
    }
    svg.appendChild(g);
    el.appendChild(svg);
    el.title = text;
    if (utf8Len(text) > 17) {
      var warn = document.createElement('p');
      warn.className = 'qr-warn';
      warn.textContent = 'Texto > 17 bytes: el QR se truncó (versión 1).';
      el.appendChild(warn);
    }
  } catch (e) {
    el.textContent = 'QR: ' + text;
  }
}

// utf8Len cuenta los bytes UTF-8 de un texto.
function utf8Len(s) {
  var n = 0;
  for (var i = 0; i < s.length; i++) {
    var c = s.charCodeAt(i);
    if (c < 0x80) n += 1;
    else if (c < 0x800) n += 2;
    else n += 3;
  }
  return n;
}

// ---- Generador QR compacto (v1, byte mode, EC-L, mask 0) ----
var EXP = [], LOG = [];
(function () {
  var x = 1;
  for (var i = 0; i < 255; i++) { EXP[i] = x; LOG[x] = i; x = (x << 1) ^ (x & 0x80 ? 0x11d : 0); }
  EXP[255] = EXP[0];
})();
function gfMul(a, b) { if (a === 0 || b === 0) return 0; return EXP[(LOG[a] + LOG[b]) % 255]; }

function rsEncode(data, ecCount) {
  var gen = [1];
  for (var i = 0; i < ecCount; i++) {
    var ng = new Array(gen.length + 1).fill(0);
    for (var j = 0; j < gen.length; j++) { ng[j] ^= gen[j]; ng[j + 1] ^= gfMul(gen[j], EXP[i]); }
    gen = ng;
  }
  var msg = data.slice().concat(new Array(ecCount).fill(0));
  for (var i = 0; i < data.length; i++) {
    var coef = msg[i];
    if (coef !== 0) {
      for (var j = 0; j < gen.length; j++) { msg[i + j] ^= gfMul(gen[j], coef); }
    }
  }
  return msg.slice(data.length);
}

function buildQR(text) {
  // byte mode, v1: capacidad 17 bytes (EC-L)
  var bytes = [];
  for (var i = 0; i < text.length; i++) {
    var c = text.charCodeAt(i);
    if (c < 0x80) bytes.push(c);
    else if (c < 0x800) { bytes.push(0xc0 | (c >> 6)); bytes.push(0x80 | (c & 0x3f)); }
    else { bytes.push(0xe0 | (c >> 12)); bytes.push(0x80 | ((c >> 6) & 0x3f)); bytes.push(0x80 | (c & 0x3f)); }
  }
  if (bytes.length > 17) bytes = bytes.slice(0, 17);

  // bits
  var bits = [];
  function pushBits(val, n) { for (var i = n - 1; i >= 0; i--) bits.push((val >> i) & 1); }
  pushBits(0x4, 4); // byte mode
  pushBits(bytes.length, 8);
  for (var i = 0; i < bytes.length; i++) pushBits(bytes[i], 8);
  for (var i = 0; i < 4 && bits.length < 17 * 8; i++) bits.push(0);
  while (bits.length % 8 !== 0) bits.push(0);
  var padBytes = [0xec, 0x11];
  var pi = 0;
  while (bits.length < 17 * 8) { pushBits(padBytes[pi % 2], 8); pi++; }

  var dataBytes = [];
  for (var i = 0; i < bits.length; i += 8) {
    var b = 0;
    for (var j = 0; j < 8; j++) b = (b << 1) | (bits[i + j] || 0);
    dataBytes.push(b);
  }
  var ec = rsEncode(dataBytes, 7);
  var codewords = dataBytes.concat(ec);

  var size = 21;
  var matrix = [];
  for (var i = 0; i < size; i++) { matrix[i] = new Array(size).fill(null); }

  // Finder patterns
  function finder(r, c) {
    for (var dr = 0; dr < 7; dr++)
      for (var dc = 0; dc < 7; dc++) {
        var v = (dr === 0 || dr === 6 || dc === 0 || dc === 6 || (dr >= 2 && dr <= 4 && dc >= 2 && dc <= 4)) ? 1 : 0;
        matrix[r + dr][c + dc] = v;
      }
  }
  finder(0, 0); finder(0, size - 7); finder(size - 7, 0);

  // Separators
  for (var i = 0; i < 8; i++) {
    if (matrix[7][i] === null) matrix[7][i] = 0;
    if (matrix[i][7] === null) matrix[i][7] = 0;
    if (matrix[7][size - 8 + i] === null) matrix[7][size - 8 + i] = 0;
    if (matrix[i][size - 8] === null) matrix[i][size - 8] = 0;
    if (matrix[size - 8][i] === null) matrix[size - 8][i] = 0;
    if (matrix[size - 8 + i][7] === null) matrix[size - 8 + i][7] = 0;
  }

  // Timing
  for (var i = 8; i < size - 8; i++) {
    matrix[6][i] = (i % 2 === 0) ? 1 : 0;
    matrix[i][6] = (i % 2 === 0) ? 1 : 0;
  }

  // Dark module
  matrix[size - 8][8] = 1;

  // Format info (EC-L, mask 0): 111011111000100
  var fmt = [1,1,1,0,1,1,1,1,1,0,0,0,1,0,0];
  // Location 1
  var loc1 = [[8,0],[8,1],[8,2],[8,3],[8,4],[8,5],[8,7],[8,8],[7,8],[5,8],[4,8],[3,8],[2,8],[1,8],[0,8]];
  for (var i = 0; i < 15; i++) matrix[loc1[i][0]][loc1[i][1]] = fmt[i];
  // Location 2 (redundant, bottom-left)
  var loc2 = [];
  for (var i = 0; i < 7; i++) loc2.push([size - 1 - i, 8]);
  loc2.push([8, size - 7]);
  for (var i = 0; i < 7; i++) loc2.push([8, size - 1 - i]);
  for (var i = 0; i < 15; i++) matrix[loc2[i][0]][loc2[i][1]] = fmt[i];

  // Place data bits in zigzag
  var bitStream = [];
  for (var i = 0; i < codewords.length; i++)
    for (var j = 7; j >= 0; j--) bitStream.push((codewords[i] >> j) & 1);

  var col = size - 1;
  var dir = -1;
  var bitI = 0;
  while (col >= 0) {
    if (col === 6) col--;
    for (var r = 0; r < size; r++) {
      var rr = (dir === -1) ? (size - 1 - r) : r;
      for (var cc = 0; cc < 2; cc++) {
        var c = col - cc;
        if (c < 0) continue;
        if (matrix[rr][c] !== null) continue;
        matrix[rr][c] = (bitI < bitStream.length) ? bitStream[bitI++] : 0;
      }
    }
    col -= 2;
    dir = -dir;
  }

  // Fill remaining nulls
  for (var y = 0; y < size; y++)
    for (var x = 0; x < size; x++)
      if (matrix[y][x] === null) matrix[y][x] = 0;

  // Apply mask 0: flip if (row+col)%2==0, except reserved
  for (var y = 0; y < size; y++) {
    for (var x = 0; x < size; x++) {
      var reserved = false;
      if (x < 9 && y < 9) reserved = true;
      if (x > size - 9 && y < 9) reserved = true;
      if (x < 9 && y > size - 9) reserved = true;
      if (x === 6 || y === 6) reserved = true;
      if (x === 8 && (y < 9 || y > size - 9)) reserved = true;
      if (y === 8 && (x < 9 || x > size - 9)) reserved = true;
      if (x === size - 8 && y === 8) reserved = true;
      if (!reserved && ((x + y) % 2 === 0)) matrix[y][x] ^= 1;
    }
  }
  return matrix;
}
