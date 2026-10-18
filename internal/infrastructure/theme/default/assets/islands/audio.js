// Isla audio: reproductor de audio con UI nativa.
export function mount(el, props) {
  var src = props.src || '';
  var title = props.title || 'Audio';
  el.classList.add('audio-player');
  el.innerHTML = '';
  if (title) { var t = document.createElement('div'); t.className = 'audio-title'; t.textContent = title; el.appendChild(t); }
  if (!src) {
    var note = document.createElement('p');
    note.className = 'audio-note';
    note.textContent = 'Requiere un archivo de audio (.mp3/.ogg/.wav).';
    el.appendChild(note);
    return;
  }
  var audio = document.createElement('audio');
  audio.controls = true;
  audio.src = src;
  audio.preload = 'metadata';
  el.appendChild(audio);
}
