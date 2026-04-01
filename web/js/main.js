import htmx from 'htmx.org';
import { initThemeToggle } from './theme-toggle.js';
import { initTransfer } from './transfer.js';
import { initPlaylist } from './playlist.js';

window.htmx = htmx;

document.addEventListener('DOMContentLoaded', function () {
  initThemeToggle();
  initTransfer();
  initPlaylist();
});
