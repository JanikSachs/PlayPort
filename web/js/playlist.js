// Handles the Transfer button in playlist-list.html.
// Uses event delegation so it works for HTMX-injected content.
// The button carries data-target-provider and data-target-input attributes
// (set in the template) to avoid relying on dynamic IDs in JS.
export function initPlaylist() {
  document.addEventListener('click', function (e) {
    const btn = e.target.closest('button[data-target-provider]');
    if (!btn) return;

    const select = document.getElementById(btn.dataset.targetProvider);
    const input = document.getElementById(btn.dataset.targetInput);
    if (!select || !input) return;

    input.value = select.value;
    if (!select.value) {
      e.preventDefault();
    }
  });
}
