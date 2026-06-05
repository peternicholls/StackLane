// app.js — Interactive mockup viewer

(function () {
    const main = document.getElementById('mainContent');
    const navTabs = document.getElementById('navTabs');
    const widthSelect = document.getElementById('widthSelect');
    const lightToggle = document.getElementById('lightModeToggle');
    let activeScreen = null;

    // ── Build nav tabs ──────────────────────────────────

    function buildNav() {
        navTabs.innerHTML = '';
        const { groups, order } = getScreenList();
        for (const group of order) {
            const label = document.createElement('span');
            label.className = 'nav-tab';
            label.textContent = group;
            label.style.fontWeight = '700';
            label.style.color = '#6272a4';
            label.style.cursor = 'default';
            label.style.fontSize = '11px';
            label.style.textTransform = 'uppercase';
            label.style.letterSpacing = '0.5px';
            navTabs.appendChild(label);

            for (const screen of groups[group]) {
                const btn = document.createElement('button');
                btn.className = 'nav-tab';
                btn.textContent = screen.title;
                btn.dataset.screen = screen.id;
                btn.addEventListener('click', () => showScreen(screen.id));
                navTabs.appendChild(btn);
            }
        }
    }

    // ── Render a screen ─────────────────────────────────

    function showScreen(id) {
        activeScreen = id;
        const screen = Screens[id];
        if (!screen) return;

        const w = parseInt(widthSelect.value, 10);
        document.documentElement.style.setProperty('--term-max-w', w + 'ch');

        // Update active tab
        navTabs.querySelectorAll('button').forEach(btn => {
            btn.classList.toggle('active', btn.dataset.screen === id);
        });

        // Build content
        const termTitle = screen.group + ' → ' + screen.title;
        let html = '';

        // Annotation callout
        if (screen.annotations) {
            html += `<div class="surface-annotations">
        <h3>${screen.annotations.heading}</h3>
        <ul>${screen.annotations.notes.map(n => `<li>${n}</li>`).join('')}</ul>
      </div>`;
        }

        // Terminal window
        html += `<div class="screen-label">${esc(termTitle)}</div>`;
        html += `<div class="terminal-window">
      <div class="terminal-titlebar">
        <span class="terminal-dot red"></span>
        <span class="terminal-dot yellow"></span>
        <span class="terminal-dot green"></span>
        <span class="terminal-title">stage — ${esc(screen.title)}</span>
        <span></span>
      </div>
      <div class="terminal-body">${typeof screen.html === 'function' ? screen.html(w) : screen.html}</div>
    </div>`;

        main.innerHTML = html;
    }

    // ── Event listeners ─────────────────────────────────

    widthSelect.addEventListener('change', () => {
        if (activeScreen) showScreen(activeScreen);
    });

    lightToggle.addEventListener('change', () => {
        document.body.classList.toggle('light', lightToggle.checked);
    });

    // ── Keyboard navigation ─────────────────────────────

    document.addEventListener('keydown', (e) => {
        const { groups, order } = getScreenList();
        const allScreens = order.flatMap(g => groups[g].map(s => s.id));
        const idx = allScreens.indexOf(activeScreen);

        if (e.key === 'ArrowRight' || e.key === 'ArrowDown') {
            e.preventDefault();
            const next = allScreens[(idx + 1) % allScreens.length];
            showScreen(next);
        } else if (e.key === 'ArrowLeft' || e.key === 'ArrowUp') {
            e.preventDefault();
            const prev = allScreens[(idx - 1 + allScreens.length) % allScreens.length];
            showScreen(prev);
        }
    });

    // ── Init ─────────────────────────────────────────────

    function esc(s) {
        return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
    }

    buildNav();
    showScreen('guided-ready');
})();
