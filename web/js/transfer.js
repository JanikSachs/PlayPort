// Handles the "Load Playlists" button in transfer.html.
// Sets hx-get with the selected provider as a query param before HTMX fires.
export function initTransfer() {
  const btn = document.getElementById('load-playlists-btn');
  if (!btn) return;

  btn.addEventListener('click', function () {
    const provider = document.getElementById('source-provider').value;
    this.setAttribute('hx-get', '/api/playlists?provider=' + encodeURIComponent(provider));
  });
}
