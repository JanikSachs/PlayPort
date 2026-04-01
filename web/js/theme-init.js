// Runs synchronously in <head> to apply dark mode before paint, preventing flash.
(function () {
  var t = localStorage.getItem('playport-theme');
  if (t === 'dark') {
    document.documentElement.setAttribute('data-theme', 'dark');
  }
})();
