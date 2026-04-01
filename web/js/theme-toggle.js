export function initThemeToggle() {
  const icon = document.getElementById('theme-icon');
  const btn = document.getElementById('theme-toggle');
  if (!btn || !icon) return;

  const theme = localStorage.getItem('playport-theme') || 'light';
  icon.textContent = theme === 'dark' ? '☀️' : '🌙';

  btn.addEventListener('click', function () {
    const current = document.documentElement.getAttribute('data-theme');
    const next = current === 'dark' ? 'light' : 'dark';
    document.documentElement.setAttribute('data-theme', next);
    icon.textContent = next === 'dark' ? '☀️' : '🌙';
    localStorage.setItem('playport-theme', next);
  });
}
