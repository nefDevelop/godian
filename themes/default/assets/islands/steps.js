// Isla steps: lista de pasos ordenados (numerados).
export function mount(el, props) {
  var items = (props.items || '').split(',').map(function (s) { return s.trim(); }).filter(Boolean);
  if (!items.length && el.textContent.trim()) {
    items = el.textContent.trim().split(/\n/).map(function (s) { return s.trim(); }).filter(Boolean);
  }
  el.classList.add('steps');
  el.innerHTML = '';
  var ol = document.createElement('ol');
  ol.className = 'steps-list';
  items.forEach(function (t, i) {
    var li = document.createElement('li');
    li.className = 'step';
    var num = document.createElement('span');
    num.className = 'step-num';
    num.textContent = i + 1;
    var txt = document.createElement('span');
    txt.className = 'step-text';
    txt.textContent = t;
    li.appendChild(num);
    li.appendChild(txt);
    ol.appendChild(li);
  });
  el.appendChild(ol);
}
