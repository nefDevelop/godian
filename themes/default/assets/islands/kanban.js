// Isla kanban: tablero con columnas y tarjetas movibles.
// Arrastrar (ratón) o botones ◀ ▶ (teclado / un solo puntero).
export function mount(el, props) {
  var columns = (props.columns || 'Por hacer,En progreso,Hecho')
    .split(',').map(function (s) { return s.trim(); }).filter(Boolean);
  el.classList.add('kanban');
  el.innerHTML = '';
  var board = document.createElement('div');
  board.className = 'kanban-board';
  columns.forEach(function (col) {
    var c = document.createElement('div');
    c.className = 'kanban-column';
    c.dataset.column = col;
    var title = document.createElement('div');
    title.className = 'kanban-column-title';
    title.textContent = col;
    c.appendChild(title);
    var list = document.createElement('div');
    list.className = 'kanban-cards';
    list.addEventListener('dragover', function (e) { e.preventDefault(); list.classList.add('is-dragover'); });
    list.addEventListener('dragleave', function () { list.classList.remove('is-dragover'); });
    list.addEventListener('drop', function (e) {
      e.preventDefault();
      list.classList.remove('is-dragover');
      var id = e.dataTransfer.getData('text/plain');
      var card = board.querySelector('[data-kanban-card="' + id + '"]');
      if (card) list.appendChild(card);
    });
    c.appendChild(list);
    board.appendChild(c);
  });

  // moveCard mueve la tarjeta a la columna anterior/siguiente (alternativa al drag).
  function moveCard(card, dir) {
    var lists = Array.prototype.slice.call(board.querySelectorAll('.kanban-cards'));
    var i = lists.indexOf(card.parentElement);
    var target = lists[i + dir];
    if (target) { target.appendChild(card); card.focus(); }
  }

  var cards = (props.cards || '').split(',').map(function (s) { return s.trim(); }).filter(Boolean);
  cards.forEach(function (t, i) {
    var card = document.createElement('div');
    card.className = 'kanban-card';
    card.tabIndex = 0;
    card.draggable = true;
    card.dataset.kanbanCard = 'c' + i;
    card.addEventListener('dragstart', function (e) { e.dataTransfer.setData('text/plain', 'c' + i); });

    var text = document.createElement('span');
    text.className = 'kanban-card-text';
    text.textContent = t;

    var actions = document.createElement('span');
    actions.className = 'kanban-card-actions';
    var left = document.createElement('button');
    left.type = 'button';
    left.className = 'kanban-move';
    left.setAttribute('aria-label', 'Mover a la columna anterior');
    left.textContent = '◀';
    left.addEventListener('click', function () { moveCard(card, -1); });
    var right = document.createElement('button');
    right.type = 'button';
    right.className = 'kanban-move';
    right.setAttribute('aria-label', 'Mover a la columna siguiente');
    right.textContent = '▶';
    right.addEventListener('click', function () { moveCard(card, 1); });
    actions.appendChild(left);
    actions.appendChild(right);

    card.appendChild(text);
    card.appendChild(actions);
    var firstList = board.querySelector('.kanban-cards');
    if (firstList) firstList.appendChild(card);
  });
  el.appendChild(board);
}
