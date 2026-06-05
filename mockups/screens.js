// screens.js — All StageServe UI surfaces rendered as terminal mockup HTML

const Screens = {};

function T(cls, text) {
  return '<span class="tok-' + cls + '">' + text.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;') + '</span>';
}

function esc(s) {
  return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
}

function rule(n) { return T('rule', '\u2500'.repeat(n)); }

function fit(text, width) {
  text = String(text);
  if (text.length >= width) return text.slice(0, Math.max(0, width));
  return text + ' '.repeat(width - text.length);
}

function pair(label, value, width, tone) {
  var labelText = fit(label, 14);
  var valueText = String(value);
  var gap = Math.max(1, width - 18 - valueText.length);
  return '  ' + T('muted', labelText) + ' ' + T(tone || 'label', valueText) + ' '.repeat(gap);
}

function summaryCards(items, width) {
  var out = [];
  for (var i = 0; i < items.length; i++) {
    var item = items[i];
    var label = fit(item.label, 18);
    out.push('  ' + T(item.tone || 'chip', item.kicker) + ' ' + T('label', label) + ' ' + T('muted', item.detail));
  }
  return out;
}

function actionLine(active, label, desc, key) {
  var marker = active ? T('active-marker', '\u25b6') : T('inactive-marker', ' ');
  var labelToken = active ? T('focus', label) : T('label', label);
  var keyText = key ? ' ' + T('subtle-chip', '[' + key + ']') : '';
  var line = marker + '  ' + labelToken + keyText;
  return [
    '  ' + (active ? '<span class="tok-cursor-item">' + line + ' '.repeat(Math.max(1, 32 - label.length)) + '</span>' : line),
    '     ' + T('muted', desc),
  ];
}

function commandLine(command, desc) {
  return '  ' + T('command', fit(command, 24)) + ' ' + T('muted', desc);
}

function panel(title, lines, width, tone) {
  tone = tone || 'neutral';
  var inner = width;
  var top = T('rule', '╭─ ' + title + ' ' + '─'.repeat(Math.max(2, inner - title.length - 5)) + '╮');
  var body = lines.map(function (line) {
    return T('rule', '│') + ' ' + line + ' '.repeat(Math.max(0, inner - stripTags(line).length - 2)) + T('rule', '│');
  });
  var bottom = T('rule', '╰' + '─'.repeat(inner) + '╯');
  return '<div class="tui-panel ' + tone + '">' + [top].concat(body).concat([bottom]).join('\n') + '</div>';
}

function popup(title, lines, width, tone) {
  tone = tone || 'neutral';
  var popupWidth = Math.min(68, Math.max(44, width - 18));
  var indent = Math.max(2, Math.floor((width - popupWidth) / 2));
  var pad = ' '.repeat(indent);
  return '<div class="popup-panel ' + tone + '">' + panel(title, lines, popupWidth, tone).split('\n').map(function (line) {
    return pad + line;
  }).join('\n') + '</div>';
}

function stripTags(html) {
  return String(html).replace(/<[^>]*>/g, '');
}

function stageHeader(surfaceLabel, width) {
  var band = width + 4;
  var title = '  StageServe';
  var right = surfaceLabel + '  local hosting console';
  var gap = Math.max(2, band - title.length - right.length);
  return '<span class="product-header">' + T('title', title) + ' '.repeat(gap) + T('surface', right) + '</span>';
}

function dashboard(items, width) {
  var gap = '  ';
  var out = [];
  var cellWidth = Math.floor((width + 4 - gap.length) / 2);
  for (var i = 0; i < items.length; i += 2) {
    var left = items[i];
    var right = items[i + 1];
    var leftText = T(left.tone || 'chip', '●') + ' ' + T('muted', fit(left.label + ': ' + left.value, cellWidth - 2));
    if (right) {
      var rightText = T(right.tone || 'chip', '●') + ' ' + T('muted', fit(right.label + ': ' + right.value, cellWidth - 2));
      out.push('<span class="dashboard-line">' + leftText + gap + rightText + '</span>');
    } else {
      out.push('<span class="dashboard-line">' + leftText + '</span>');
    }
  }
  return out.join('\n');
}

function commandStrip(commands, width) {
  var text = commands.map(function (cmd) {
    return T('hotkey', cmd.key) + T('muted', ' ' + cmd.label);
  }).join('   ');
  return [
    T('rule', '─'.repeat(width + 4)),
    text,
    T('rule', '─'.repeat(width + 4)),
  ].join('\n');
}

function productFooter(width) {
  var text = '(c) StageServe 2026  |  Help: stage doctor  |  Docs: README  |  Contact: team';
  return [
    T('rule', '━'.repeat(width + 4)),
    T('footer', fit(text, width + 4)),
  ].join('\n');
}

function chrome(surfaceLabel, width, statusItems, commands) {
  return [
    stageHeader(surfaceLabel, width),
    dashboard(statusItems, width),
    commandStrip(commands, width),
  ].join('\n');
}

function section(label, tone, width) {
  tone = tone || 'neutral';
  var fill = width - 3 - label.length - 1;
  if (fill < 2) fill = 2;
  return '<span class="section-title ' + tone + '"><span class="st-dash">\u2500\u2500 </span><span class="st-label">' + esc(label) + '</span><span class="st-dash"> ' + '\u2500'.repeat(fill) + '</span></span>';
}

function hdr(surfaceLabel, width) {
  var left = T('accent', '\u25c6') + '  ' + T('title', 'StageServe');
  if (!surfaceLabel) return '  ' + left;
  var leftW = 15;
  var gap = Math.max(2, width - leftW - surfaceLabel.length);
  return '  ' + left + ' '.repeat(gap) + T('surface', surfaceLabel);
}

// 1. GUIDED SHELL — Project ready to run
Screens['guided-ready'] = {
  group: 'Guided Shell',
  title: 'Project ready to run',
  annotations: {
    heading: 'Surface: Guided shell \u2014 ready state',
    notes: [
      'The default landing surface when a project has <code>.env.stageserve</code> but isn\'t running',
      'Shows <code>Project snapshot</code>, <code>Health checks</code>, and primary actions in a stronger hierarchy',
      'Adds shortcut hints for setup, edit, diagnostics, and advanced utility access',
      'Navigation keys sit above content; the bottom footer is reserved for quiet help and product information',
      'Code: <code>core/guidance/shell.go \u2192 renderShellViewState</code>',
    ]
  },
  html: function (w) {
    var lw = w - 4;
    return [
      chrome('Ready', lw, [
        { tone: 'ready', label: 'Machine ready', value: 'Docker and ports passed' },
        { tone: 'ready', label: 'DNS ready', value: '*.test points here' },
        { tone: 'subtle-chip', label: 'Project stopped', value: 'Configured, not running' },
        { tone: 'chip', label: 'Local URL', value: 'https://demo.test' },
      ], [
        { key: '↵', label: 'run project' },
        { key: 'e', label: 'edit settings' },
        { key: 'd', label: 'diagnostics' },
        { key: 'm', label: 'more' },
      ]),
      '',
      '  ' + T('verdict-ready', 'This project is ready to run.'),
      '  ' + T('muted', 'Nothing is running yet. Start the local site when you are ready.'),
      '',
      section('Project snapshot', 'neutral', lw),
      pair('Site name', 'demo', lw, 'label'),
      pair('Local URL', 'https://demo.test', lw, 'command'),
      pair('Web folder', 'public_html', lw, 'path'),
      pair('Stack', '20i shared hosting', lw, 'label'),
      '',
      section('Health checks', 'action', lw),
    ].concat(summaryCards([
      { kicker: '\u2713', tone: 'ready', label: 'Machine ready', detail: 'Docker and ports passed' },
      { kicker: '\u2713', tone: 'ready', label: 'DNS ready', detail: '*.test resolves locally' },
      { kicker: '\u25cf', tone: 'subtle-chip', label: 'Project stopped', detail: 'No services active' },
    ], lw)).concat([
      '',
      section('Primary actions', 'action', lw),
    ]).concat(actionLine(true, 'Run this project', 'Start the local site and open it in your browser.', '↵')).concat([
      ''
    ]).concat(actionLine(false, 'Edit project settings', 'Change the site name, web folder, or local address.', 'e')).concat([
      ''
    ]).concat(actionLine(false, 'Run diagnostics', 'Check machine and project readiness without changing anything.', 'd')).concat([
      '',
      productFooter(lw),
      '',
    ]).join('\n');
  }
};

// 2. GUIDED SHELL — Running project
Screens['guided-running'] = {
  group: 'Guided Shell',
  title: 'Project running',
  annotations: {
    heading: 'Surface: Guided shell \u2014 running state',
    notes: [
      'The main surface when the project is actively running',
      'Shows expanded action set: logs, browser, status, restart, stop',
      'More... action expands to show direct commands and troubleshooting',
      'Code: <code>core/guidance/planner.go \u2192 SituationProjectRunning</code>',
    ]
  },
  html: function (w) {
    var lw = w - 4;
    return [
      chrome('Running', lw, [
        { tone: 'ready', label: 'Site online', value: 'Browser can open demo.test' },
        { tone: 'ready', label: 'DNS ready', value: '*.test routes locally' },
        { tone: 'ready', label: 'Services running', value: 'apache and nginx are up' },
        { tone: 'chip', label: 'Stack', value: '20i shared hosting' },
      ], [
        { key: 'o', label: 'open browser' },
        { key: 'l', label: 'logs' },
        { key: 's', label: 'status' },
        { key: 'm', label: 'more' },
      ]),
      '',
      '  ' + T('verdict-ready', 'This project is running at https://demo.test.'),
      '  ' + T('muted', 'The local site is available. Keep the common controls close at hand.'),
      '',
      panel('Workspace', [
        T('ready', '●') + ' ' + T('label', 'demo.test') + '        ' + T('muted', 'online and routed through local DNS'),
        T('ready', '●') + ' ' + T('label', 'apache') + '           ' + T('muted', 'running  ·  PHP enabled  ·  last restart 3h ago'),
        T('chip', '●') + ' ' + T('label', 'public_html') + '      ' + T('muted', 'serving project files from the configured web folder'),
      ], lw, 'ready'),
      '',
      section('Live status', 'neutral', lw),
    ].concat(summaryCards([
      { kicker: '\u25cf', tone: 'ready', label: 'Local site online', detail: 'https://demo.test' },
      { kicker: '\u2713', tone: 'ready', label: 'Apache + PHP', detail: 'Up 3 hours' },
      { kicker: '\u25b2', tone: 'chip', label: 'URL routing', detail: '*.test to localhost' },
      { kicker: '2', tone: 'subtle-chip', label: 'Services active', detail: 'apache, nginx' },
    ], lw)).concat([
      '',
      section('Project controls', 'action', lw),
    ]).concat(actionLine(true, 'Open in browser', 'Open https://demo.test in your default browser.', 'o')).concat([
      ''
    ]).concat(actionLine(false, 'View project logs', 'Watch recent project activity and errors.', 'l')).concat([
      ''
    ]).concat(actionLine(false, "Check this project's status", 'See live service status and routing details.', 's')).concat([
      ''
    ]).concat(actionLine(false, 'Restart services', 'Restart the project services after confirmation.', 'r')).concat([
      ''
    ]).concat(actionLine(false, 'Stop this project', 'Stop the local site after confirmation.', 'x')).concat([
      '',
      section('Utilities', 'neutral', lw),
      commandLine('stage doctor', 'Read-only setup checks'),
      commandLine('stage --cli', 'Plain text guided output'),
      commandLine('stage logs apache', 'Direct logs command'),
      '',
      productFooter(lw),
      '',
    ]).join('\n');
  }
};

// 3. GUIDED SHELL — Loading state
Screens['guided-loading'] = {
  group: 'Guided Shell',
  title: 'Loading (action in progress)',
  annotations: {
    heading: 'Surface: Loading view \u2014 async action progress',
    notes: [
      'Shown while an async action runs (up, down, restart, etc.)',
      'Verdict-first: shows what\'s happening at the top',
      '<code>Current step</code> section shows spinner + detail text',
      'Code: <code>core/guidance/shell.go \u2192 renderLoadingView</code>',
    ]
  },
  html: function (w) {
    var lw = w - 4;
    return [
      chrome('Starting up', lw, [
        { tone: 'chip', label: 'Action running', value: 'Starting local services' },
        { tone: 'ready', label: 'DNS ready', value: '*.test routes locally' },
        { tone: 'subtle-chip', label: 'Project', value: 'demo' },
        { tone: 'ready', label: 'Config ready', value: '.env.stageserve found' },
      ], [
        { key: 'esc', label: 'cancel' },
        { key: 'q', label: 'quit after action' },
      ]),
      '',
      '  ' + T('verdict', 'Starting this project...'),
      '',
      section('Current step', 'action', lw),
      '  <span class="spinner">\u23fe</span> ' + T('label', 'In progress'),
      '    ' + T('muted', 'Bringing project containers online'),
      '',
      productFooter(lw),
      '',
    ].join('\n');
  }
};

// 4. GUIDED SHELL — Confirmation (safe)
Screens['guided-confirm-safe'] = {
  group: 'Guided Shell',
  title: 'Confirmation (safe action)',
  annotations: {
    heading: 'Surface: Confirmation modal \u2014 safe action',
    notes: [
      'Bordered panel appears inline when an action needs confirmation',
      'Safe actions (init, up, attach) use a dim/cyan border',
      'Shows action title, label, description, and summary',
      'Code: <code>core/guidance/shell.go \u2192 renderConfirmationModal</code>',
    ]
  },
  html: function (w) {
    var lw = w - 4;
    return [
      hdr('Ready', lw),
      '  ' + rule(lw),
      '',
      '  ' + T('verdict-ready', 'This project is ready to run.'),
      '',
      section('What you can do', 'action', lw),
      '',
      '  <span class="tok-cursor-item">' + T('active-marker', '\u25b6') + '  ' + T('focus', 'Run this project') + '                   </span>',
      '     Start the project and open it in your browser.',
      '',
      '  ' + T('inactive-marker', ' ') + '  ' + T('label', 'Edit project settings') + '                  ',
      '     Change the project name, web folder, or local address first.',
      '',
      '<div class="confirm-panel create">',
      '  ' + T('label', 'Confirm change'),
      '  ' + T('muted', 'Run this project'),
      '  ' + T('muted', 'Start the project and open it in your browser.'),
      '',
      '  ' + T('muted', 'Applies the selected change to demo after confirmation.'),
      '',
      '  ' + T('footer', '[enter] confirm  [y] confirm  [esc] cancel'),
      '</div>',
      '',
      '  ' + rule(lw),
      '  ' + T('footer', '\u2191\u2193 navigate  \u21b5 select  d details  q quit'),
      '',
    ].join('\n');
  }
};

// 5. GUIDED SHELL — Confirmation (destructive)
Screens['guided-confirm-destructive'] = {
  group: 'Guided Shell',
  title: 'Confirmation (destructive)',
  annotations: {
    heading: 'Surface: Confirmation modal \u2014 destructive action',
    notes: [
      'Destructive actions (down, detach) use a RED border',
      'Title changes to "Confirm stop"',
      'Extra warning: "Default: cancel. Press y only if you want this change."',
    ]
  },
  html: function (w) {
    var lw = w - 4;
    return [
      hdr('Running', lw),
      '  ' + rule(lw),
      '',
      '  ' + T('verdict-ready', 'This project is running at https://demo.test.'),
      '',
      section('What you can do', 'action', lw),
      '',
      '  ' + T('inactive-marker', ' ') + '  View project logs                       ',
      '',
      '  ' + T('inactive-marker', ' ') + '  Open in browser                         ',
      '',
      '  ' + T('inactive-marker', ' ') + "  Check this project's status              ",
      '',
      '  ' + T('inactive-marker', ' ') + '  Restart apache                           ',
      '',
      '  <span class="tok-cursor-item">' + T('active-marker', '\u25b6') + '  ' + T('focus', 'Stop this project') + '             </span>',
      '     Stop the project after confirmation.',
      '',
      '<div class="confirm-panel destructive">',
      '  ' + T('label', 'Confirm stop'),
      '  ' + T('muted', 'Stop this project'),
      '  ' + T('muted', 'Stop the project after confirmation.'),
      '',
      '  ' + T('muted', 'Stops https://demo.test; project files and settings are not changed.'),
      '',
      '  ' + T('muted', 'Default: cancel. Press y only if you want this change.'),
      '',
      '  ' + T('footer', '[y] confirm  [enter] cancel  [esc] cancel'),
      '</div>',
      '',
      '  ' + rule(lw),
      '  ' + T('footer', '\u2191\u2193 navigate  \u21b5 select  d details  q quit'),
      '',
    ].join('\n');
  }
};

// 6. GUIDED SHELL — Machine not ready
Screens['guided-machine-not-ready'] = {
  group: 'Guided Shell',
  title: 'Machine not ready',
  annotations: {
    heading: 'Surface: Guided shell \u2014 machine readiness blocking',
    notes: [
      'Shown when Docker, ports, DNS, or other machine-level prereqs aren\'t met',
      'Uses YELLOW verdict tone',
      '<code>Setup steps</code> shows work items with markers',
      'Work items appear BEFORE decision items',
      'Code: <code>core/guidance/planner.go \u2192 SituationMachineNotReady</code>',
    ]
  },
  html: function (w) {
    var lw = w - 4;
    return [
      chrome('Needs attention', lw, [
        { tone: 'needs-action', label: 'Docker needs action', value: 'Open Docker Desktop first' },
        { tone: 'subtle-chip', label: 'DNS not checked', value: 'Waiting for Docker' },
        { tone: 'subtle-chip', label: 'Ports not checked', value: 'Waiting for setup' },
        { tone: 'subtle-chip', label: 'Project pending', value: 'Next step after machine is ready' },
      ], [
        { key: '↵', label: 'open Docker' },
        { key: 'd', label: 'diagnostics' },
        { key: 'm', label: 'setup commands' },
        { key: 'q', label: 'quit' },
      ]),
      '',
      '  ' + T('verdict-warn', "Your computer isn't ready yet."),
      '  ' + T('muted', 'Fix the next setup item, then StageServe can plan the project step.'),
      '',
      section('Readiness summary', 'warning', lw),
      '  ' + T('ready', '\u2713') + '  ' + T('label', fit('Docker CLI', 26)) + T('dim', 'installed'),
      '  ' + T('needs-action', '\u25b6') + '  ' + T('label', fit('Docker Desktop', 26)) + T('muted', 'not running'),
      '  ' + T('dim', '\u2022') + '  ' + T('dim', fit('Port 80', 26)) + T('dim', 'waiting'),
      '  ' + T('dim', '\u2022') + '  ' + T('dim', fit('Port 443', 26)) + T('dim', 'waiting'),
      '  ' + T('dim', '\u2022') + '  ' + T('dim', fit('Local URL routing', 26)) + T('dim', 'waiting'),
      '',
      section('Next best steps', 'action', lw),
    ].concat(actionLine(true, 'Open Docker Desktop', 'Start Docker, then continue setup checks.', '↵')).concat([
      '',
    ]).concat(actionLine(false, 'Run diagnostics', 'Read-only. Check machine and project readiness.', 'd')).concat([
      ''
    ]).concat(actionLine(false, 'Show setup commands', 'Use direct commands if you prefer the shell path.', 'm')).concat([
      '',
      '',
      productFooter(lw),
      '',
    ]).join('\n');
  }
};

// 7. GUIDED SHELL — Project settings editor
Screens['guided-editor'] = {
  group: 'Guided Shell',
  title: 'Project settings editor',
  annotations: {
    heading: 'Surface: Inline settings editor',
    notes: [
      'Replaces the decision section with a field editor',
      '\u25b6 marker indicates the active field (cycles with tab)',
      'Local URL is computed read-only from site name + suffix',
      'Code: <code>core/guidance/shell.go \u2192 renderProjectSettingsEditor</code>',
    ]
  },
  html: function (w) {
    var lw = w - 4;
    return [
      hdr('Editing', lw),
      '  ' + rule(lw),
      '',
      '  ' + T('verdict', "This folder doesn't have StageServe settings yet."),
      '  ' + T('muted', 'StageServe can create a small .env.stageserve file for this project.'),
      '',
      section('Project settings', 'neutral', lw),
      '',
      '  ' + T('active-marker', '\u25b6') + ' ' + T('focus', 'Site name      ') + ' ' + T('focus', 'demo'),
      '  ' + T('inactive-marker', ' ') + ' ' + T('label', 'Web folder     ') + ' public_html',
      '  ' + T('inactive-marker', ' ') + ' ' + T('label', 'Domain suffix  ') + ' test',
      '',
      '  ' + T('label', 'Local URL      ') + ' https://demo.test',
      '  ' + T('muted', 'Saving updates the preview only; confirm project settings to write the file.'),
      '',
      '  ' + rule(lw),
      '  ' + T('footer', 'tab next field  \u21b5 confirm  esc cancel'),
      '',
    ].join('\n');
  }
};

// 8. GUIDED SHELL — Not a project
Screens['guided-not-project'] = {
  group: 'Guided Shell',
  title: 'Not a project directory',
  annotations: {
    heading: 'Surface: Guided shell \u2014 not a project',
    notes: [
      'Shown when running <code>stage</code> in a directory that isn\'t configured',
      'Single action: "Set up this folder as a project"',
      'Code: <code>core/guidance/planner.go \u2192 SituationNotProject</code>',
    ]
  },
  html: function (w) {
    var lw = w - 4;
    return [
      hdr('Status', lw),
      '  ' + rule(lw),
      '',
      '  ' + T('verdict', "This folder isn't a StageServe project yet."),
      '',
      section('What you can do', 'action', lw),
      '',
      '  <span class="tok-cursor-item">' + T('active-marker', '\u25b6') + '  ' + T('focus', 'Set up this folder as a project') + '     </span>',
      '     Create project settings for this folder.',
      '',
      '',
      '  ' + rule(lw),
      '  ' + T('footer', '\u2191\u2193 navigate  \u21b5 select  d details  q quit'),
      '',
    ].join('\n');
  }
};

// 9. GUIDED SHELL — Project stopped
Screens['guided-stopped'] = {
  group: 'Guided Shell',
  title: 'Project stopped',
  annotations: {
    heading: 'Surface: Guided shell \u2014 project down',
    notes: [
      'Project was previously configured but is now stopped',
      'Offers restart, status check, and destructive detach action',
      'Code: <code>core/guidance/planner.go \u2192 SituationProjectDown</code>',
    ]
  },
  html: function (w) {
    var lw = w - 4;
    return [
      hdr('Stopped', lw),
      '  ' + rule(lw),
      '',
      '  ' + T('verdict', 'This project is stopped.'),
      '',
      section('What you can do', 'action', lw),
      '',
      '  <span class="tok-cursor-item">' + T('active-marker', '\u25b6') + '  ' + T('focus', 'Run this project again') + '             </span>',
      '     Start the project and open it in your browser.',
      '',
      '  ' + T('inactive-marker', ' ') + '  ' + T('label', "Check this project's status") + '         ',
      '     See the latest recorded and live project status.',
      '',
      '  ' + T('inactive-marker', ' ') + '  Remove this project from StageServe    ',
      '     Stop tracking this project. Your files will not be touched.',
      '',
      '',
      '  ' + rule(lw),
      '  ' + T('footer', '\u2191\u2193 navigate  \u21b5 select  d details  q quit'),
      '',
    ].join('\n');
  }
};

// 10. GUIDED SHELL — Drift detected / recovery
Screens['guided-drift'] = {
  group: 'Guided Shell',
  title: 'Drift detected / recovery',
  annotations: {
    heading: 'Surface: Guided shell \u2014 drift / error recovery',
    notes: [
      'Shown when runtime state doesn\'t match expectations',
      'Work items (recovery steps) appear first',
      'Numbered action steps progress from safe (read-only) to risky',
      'Code: <code>core/guidance/planner.go \u2192 SituationDriftDetected</code>',
    ]
  },
  html: function (w) {
    var lw = w - 4;
    return [
      hdr('Needs attention', lw),
      '  ' + rule(lw),
      '',
      '  ' + T('verdict-warn', "This project doesn't match what StageServe expects."),
      '  ' + T('muted', 'StageServe can walk through the safest checks first.'),
      '',
      section('Recovery steps', 'warning', lw),
      '',
      '  ' + T('needs-action', '\u25b6') + '  ' + T('label', 'Run diagnostics') + '                     ' + T('muted', 'read-only'),
      '  ' + T('dim', '\u2022') + '  ' + T('dim', 'Check project status') + '                    ' + T('dim', 'not started'),
      '  ' + T('dim', '\u2022') + '  ' + T('dim', 'Look at project logs') + '                     ' + T('dim', 'not started'),
      '',
      section('What you can do', 'action', lw),
      '',
      '  <span class="tok-cursor-item">' + T('active-marker', '\u25b6') + '  ' + T('focus', 'Step 1: run diagnostics') + '          </span>',
      '     Read-only. Check machine and runtime readiness.',
      '',
      '  ' + T('inactive-marker', ' ') + '  ' + T('label', "Step 2: check this project's status") + '  ',
      '     Read-only. Nothing on your computer will be changed.',
      '',
      '  ' + T('inactive-marker', ' ') + '  Step 3: look at the latest project log  ',
      '     Read-only. This shows the latest log output for the project.',
      '',
      '  ' + T('inactive-marker', ' ') + '  Step 4: try running this project again ',
      '     Restart with the current settings.',
      '',
      '  ' + T('inactive-marker', ' ') + '  Step 5: stop this project first         ',
      '     Stop the project, then try again.',
      '',
      '',
      '  ' + rule(lw),
      '  ' + T('footer', '\u2191\u2193 navigate  \u21b5 select  d details  q quit'),
      '',
    ].join('\n');
  }
};

// 11. GUIDED SHELL — Details panel
Screens['guided-details'] = {
  group: 'Guided Shell',
  title: 'Details panel expanded',
  annotations: {
    heading: 'Surface: Details panel \u2014 toggled with d key',
    notes: [
      'Shows warnings and direct CLI commands',
      'Direct commands are rendered in bright cyan bold',
      'Code: <code>core/guidance/shell.go \u2192 renderDetailsSection</code>',
    ]
  },
  html: function (w) {
    var lw = w - 4;
    return [
      hdr('Running', lw),
      '  ' + rule(lw),
      '',
      '  ' + T('verdict-ready', 'This project is running at https://demo.test.'),
      '',
      section('What you can do', 'action', lw),
      '',
      '  <span class="tok-cursor-item">' + T('active-marker', '\u25b6') + '  ' + T('focus', 'View project logs') + '              </span>',
      '     Watch what your project is doing right now.',
      '',
      '  ' + T('inactive-marker', ' ') + '  Open in browser                         ',
      '',
      '  ' + T('inactive-marker', ' ') + "  Check this project's status              ",
      '',
      '  ' + T('inactive-marker', ' ') + '  Restart apache                           ',
      '',
      '  ' + T('inactive-marker', ' ') + '  Stop this project                        ',
      '',
      '  ' + T('inactive-marker', ' ') + '  More...                                  ',
      '',
      section('Details', 'neutral', lw),
      '',
      '  ' + T('muted', 'No extra warnings for this plan.'),
      '',
      '  ' + T('command', 'stage status'),
      '  ' + T('command', 'stage logs apache'),
      '  ' + T('command', 'stage down'),
      '',
      '  ' + rule(lw),
      '  ' + T('footer', '\u2191\u2193 navigate  \u21b5 select  d hide details  q quit'),
      '',
    ].join('\n');
  }
};

// 12. GUIDED SHELL — Doctor utility surface
Screens['guided-doctor-utility'] = {
  group: 'Guided Shell',
  title: 'Doctor report (utility surface)',
  annotations: {
    heading: 'Surface: Utility surface \u2014 doctor report in guided shell',
    notes: [
      'Replaces the entire decision section with a rich styled report',
      'Uses the TUIProjector from projection_tui.go',
      'Full colour palette: green ready, yellow warn, red error',
      'Bright cyan bold for remediation commands',
      'Footer shows "Back to menu" navigation',
    ]
  },
  html: function (w) {
    var lw = w - 4;
    return [
      hdr('StageServe Doctor', lw),
      '  ' + rule(lw),
      '',
      section('StageServe Doctor', 'neutral', lw),
      '',
      '  ' + T('ready', '\u2713') + '  ' + T('white-bold', 'All 8 checks passed \u2014 your machine is ready.'),
      '',
      section('Checks passed', 'action', lw),
      '',
      '  ' + T('ready', '\u2713') + '  ' + T('label', 'Docker CLI           ') + '  ' + T('dim', 'installed'),
      '  ' + T('ready', '\u2713') + '  ' + T('label', 'Docker daemon        ') + '  ' + T('dim', 'running'),
      '  ' + T('ready', '\u2713') + '  ' + T('label', 'State directory      ') + '  ' + T('dim', 'exists'),
      '  ' + T('ready', '\u2713') + '  ' + T('label', 'Shared stack file    ') + '  ' + T('dim', 'available'),
      '  ' + T('ready', '\u2713') + '  ' + T('label', 'Project stack file   ') + '  ' + T('dim', 'available'),
      '  ' + T('ready', '\u2713') + '  ' + T('label', 'Port 80              ') + '  ' + T('dim', 'available'),
      '  ' + T('ready', '\u2713') + '  ' + T('label', 'Port 443             ') + '  ' + T('dim', 'available'),
      '  ' + T('ready', '\u2713') + '  ' + T('label', 'DNS resolver         ') + '  ' + T('dim', 'configured'),
      '',
      '  ' + T('dim', '\u2500'.repeat(40)),
      '  ' + T('dim', 'Your machine is ready. Run:') + ' ' + T('bright-cyan', 'stage up'),
      '',
      '  ' + rule(lw),
      '  ' + T('footer', '\u21b5 back to menu  q quit'),
      '',
    ].join('\n');
  }
};

// 13. DOCTOR REPORT — With issues (CLI)
Screens['doctor-issues'] = {
  group: 'CLI Commands',
  title: 'Doctor report (issues found)',
  annotations: {
    heading: 'Surface: Doctor report \u2014 standalone CLI with issues',
    notes: [
      'The full <code>stage doctor</code> output when issues are detected',
      '"Needs fixing" section with numbered issues, descriptions, and remedies',
      'Issue colours: YELLOW for needs-action, RED for error',
      '"To fix:" in bold, command in bright cyan bold',
      'Code: <code>core/onboarding/projection_tui.go \u2192 projectDetailed</code>',
    ]
  },
  html: function (w) {
    var lw = w - 4;
    return [
      '',
      '  ' + T('accent', '\u25c6') + '  ' + T('white-bold', 'StageServe Doctor'),
      '  ' + T('dim', '\u2500'.repeat(40)),
      '',
      '  ' + T('error', '\u2717') + '  ' + T('error', 'Not ready \u2014 2 of 8 checks need attention.'),
      '',
      section('Needs fixing', 'warning', lw),
      '',
      '  ' + T('needs-action', '1') + '  ' + T('white-bold', 'Docker daemon'),
      '     ' + T('italic', 'The Docker daemon must be running before any container can start.'),
      '',
      '     ' + T('dim', 'Docker is installed but not running'),
      '     ' + T('label', 'To fix:') + '  ' + T('bright-cyan', 'open Docker Desktop or run: open -a Docker'),
      '',
      '  ' + T('error', '2') + '  ' + T('white-bold', 'DNS resolver'),
      '     ' + T('italic', 'Routes *.test domains to your stack \u2014 needs dnsmasq configured.'),
      '',
      '     ' + T('dim', 'dnsmasq not configured for *.test'),
      '     ' + T('label', 'To fix:') + '  ' + T('bright-cyan', 'stage dns-setup'),
      '',
      section('All clear', 'action', lw),
      '',
      '  ' + T('ready', '\u2713') + '  ' + T('label', 'Docker CLI           ') + '  ' + T('dim', 'installed'),
      '  ' + T('ready', '\u2713') + '  ' + T('label', 'State directory      ') + '  ' + T('dim', 'exists'),
      '  ' + T('ready', '\u2713') + '  ' + T('label', 'Shared stack file    ') + '  ' + T('dim', 'available'),
      '  ' + T('ready', '\u2713') + '  ' + T('label', 'Project stack file   ') + '  ' + T('dim', 'available'),
      '  ' + T('ready', '\u2713') + '  ' + T('label', 'Port 80              ') + '  ' + T('dim', 'available'),
      '  ' + T('ready', '\u2713') + '  ' + T('label', 'Port 443             ') + '  ' + T('dim', 'available'),
      '',
      '  ' + T('dim', '\u2500'.repeat(40)),
      '  ' + T('dim', 'Fix the issues above, then run:') + ' ' + T('bright-cyan', 'stage doctor'),
      '',
    ].join('\n');
  }
};

// 14. SETUP — CLI TUI mode
Screens['setup-cli'] = {
  group: 'CLI Commands',
  title: 'stage setup (all checks pass)',
  annotations: {
    heading: 'Surface: stage setup \u2014 machine setup report',
    notes: [
      'Same projector as doctor but with Title: "StageServe Setup"',
      'Setup may auto-fix issues (unlike doctor which is read-only)',
      'Code: <code>cmd/stage/commands/setup.go \u2192 NewSetup</code>',
    ]
  },
  html: function (w) {
    return [
      '',
      '  ' + T('accent', '\u25c6') + '  ' + T('white-bold', 'StageServe Setup'),
      '  ' + T('dim', '\u2500'.repeat(40)),
      '',
      '  ' + T('ready', '\u2713') + '  ' + T('ready', 'All 8 checks passed \u2014 your machine is ready.'),
      '',
      section('Checks passed', 'action', w - 4),
      '',
      '  ' + T('ready', '\u2713') + '  ' + T('label', 'Docker CLI           ') + '  ' + T('dim', 'installed'),
      '  ' + T('ready', '\u2713') + '  ' + T('label', 'Docker daemon        ') + '  ' + T('dim', 'running'),
      '  ' + T('ready', '\u2713') + '  ' + T('label', 'State directory      ') + '  ' + T('dim', 'exists'),
      '  ' + T('ready', '\u2713') + '  ' + T('label', 'Shared stack file    ') + '  ' + T('dim', 'available'),
      '  ' + T('ready', '\u2713') + '  ' + T('label', 'Project stack file   ') + '  ' + T('dim', 'available'),
      '  ' + T('ready', '\u2713') + '  ' + T('label', 'Port 80              ') + '  ' + T('dim', 'available'),
      '  ' + T('ready', '\u2713') + '  ' + T('label', 'Port 443             ') + '  ' + T('dim', 'available'),
      '  ' + T('ready', '\u2713') + '  ' + T('label', 'DNS resolver         ') + '  ' + T('dim', 'configured'),
      '',
      '  ' + T('dim', '\u2500'.repeat(40)),
      '  ' + T('ready', 'Your machine is ready. Run:') + ' ' + T('bright-cyan', 'stage up'),
      '',
    ].join('\n');
  }
};

// 15. GUIDED SHELL — More panel
Screens['guided-more'] = {
  group: 'Guided Shell',
  title: 'More panel (commands & troubleshooting)',
  annotations: {
    heading: 'Surface: Utility surface \u2014 "More..." panel',
    notes: [
      'Triggered by selecting "More..." action from the running project menu',
      'Subheadings styled with bold label',
      'Commands styled with bright cyan bold',
      'Code: <code>core/guidance/shell.go \u2192 renderUtilitySurface</code>',
    ]
  },
  html: function (w) {
    var lw = w - 4;
    return [
      chrome('More', lw, [
        { tone: 'ready', label: 'Project running', value: 'Local site is online' },
        { tone: 'ready', label: 'DNS ready', value: '*.test routes locally' },
        { tone: 'chip', label: 'Tools available', value: 'Diagnostics and output modes' },
        { tone: 'subtle-chip', label: 'Mode advanced', value: 'Extra controls only' },
      ], [
        { key: '↵', label: 'choose tool' },
        { key: '/', label: 'search tools' },
        { key: 'esc', label: 'back to menu' },
        { key: 'q', label: 'quit' },
      ]),
      '',
      '  ' + T('verdict-ready', 'This project is running at https://demo.test.'),
      '  ' + T('muted', 'Advanced options stay here so the main screen can stay focused.'),
      '',
      section('Project tools', 'action', lw),
    ].concat(actionLine(true, 'Open in browser', 'Open the local site in your default browser.', 'o')).concat([
      '',
    ]).concat(actionLine(false, 'View logs', 'Stream recent output from the running services.', 'l')).concat([
      ''
    ]).concat(actionLine(false, 'Check status', 'Inspect service health, URL, and stack details.', 's')).concat([
      '',
      section('Advanced and troubleshooting', 'neutral', lw),
      commandLine('stage doctor', 'Read-only setup and routing checks'),
      commandLine('stage setup', 'Guided machine setup'),
      commandLine('stage status --json', 'Machine-readable project status'),
      commandLine('stage --cli', 'Plain text guided output'),
      '',
      section('Change controls', 'warning', lw),
      commandLine('stage restart apache', 'Restart one service after confirmation'),
      '  ' + T('danger-command', fit('stage down', 24)) + ' ' + T('muted', 'Stop this project after confirmation'),
      '  ' + T('danger-command', fit('stage project remove', 24)) + ' ' + T('muted', 'Remove StageServe tracking only'),
      '',
      productFooter(lw),
      '',
    ]).join('\n');
  }
};

// 15. GUIDED SHELL — Command palette popup
Screens['guided-command-palette'] = {
  group: 'Guided Shell',
  title: 'Command palette popup',
  annotations: {
    heading: 'Surface: Bubble Tea command palette',
    notes: [
      'A modal palette keeps advanced actions available without crowding the main screen',
      'Uses familiar list/search behavior: type to filter, arrows to move, enter to run, esc to dismiss',
      'This is a mockup target for Bubbles list/textinput/help composition inside the guided shell',
    ]
  },
  html: function (w) {
    var lw = w - 4;
    return [
      chrome('Running', lw, [
        { tone: 'ready', label: 'Site online', value: 'Browser can open demo.test' },
        { tone: 'ready', label: 'DNS ready', value: '*.test routes locally' },
        { tone: 'ready', label: 'Services running', value: 'apache and nginx are up' },
        { tone: 'chip', label: 'Palette open', value: 'Filtering project actions' },
      ], [
        { key: '↑↓', label: 'move' },
        { key: '↵', label: 'choose' },
        { key: '/', label: 'filter' },
        { key: 'esc', label: 'close' },
      ]),
      '',
      '  ' + T('verdict-ready', 'This project is running at https://demo.test.'),
      '  ' + T('muted', 'The command palette is open. The project keeps running behind it.'),
      '',
      popup('Command palette', [
        T('muted', 'Search') + '  ' + T('focus', 'status'),
        '',
        '<span class="tok-cursor-item">' + T('active-marker', '▶') + '  ' + T('focus', "Check this project's status") + '     ' + T('subtle-chip', 'read-only') + '</span>',
        '   ' + T('muted', 'Show service health, local URL, DNS, and stack details.'),
        '   ' + T('label', 'View project logs') + '              ' + T('subtle-chip', 'read-only'),
        '   ' + T('label', 'Restart services') + '               ' + T('subtle-chip', 'confirmation'),
        '   ' + T('label', 'Open advanced troubleshooting') + '   ' + T('subtle-chip', 'tools'),
        '',
        T('footer', '↑↓ move   ↵ run   / filter   esc close'),
      ], lw, 'palette'),
      '',
      productFooter(lw),
      '',
    ].join('\n');
  }
};

// 16. PLAIN TEXT — Ready to run
Screens['plaintext-ready'] = {
  group: 'Plain Text',
  title: 'Plain text output (--cli)',
  annotations: {
    heading: 'Surface: Plain text fallback \u2014 ready to run',
    notes: [
      'The <code>stage --cli</code> or <code>stage --no-tui</code> output',
      'No ANSI colours \u2014 pure monochrome text',
      'Code: <code>core/guidance/text.go \u2192 RenderText</code>',
    ]
  },
  html: function (w) {
    var lw = w - 4;
    return [
      'StageServe',
      '\u2500'.repeat(lw),
      'Ready',
      '',
      'This project is ready to run.',
      '',
      'Key facts:',
      '  Site name:      demo',
      '  Local URL:      https://demo.test',
      '  Web folder:     public_html',
      '',
      'What you can do:',
      '  1. Run this project',
      '     Start the project and open it in your browser.',
      '  2. Edit project settings',
      '     Change the project name, web folder, or local address.',
      '',
      '  ' + '\u2500'.repeat(Math.max(0, lw - 2)),
      '  stage up  stage status',
      '',
    ].join('\n');
  }
};

// 17. JSON output
Screens['json-output'] = {
  group: 'Plain Text',
  title: 'JSON output (--json)',
  annotations: {
    heading: 'Surface: JSON output \u2014 machine-readable envelope',
    notes: [
      'The <code>stage doctor --json</code> or <code>stage setup --json</code> output',
      'Structured JSON for CI/CD pipelines and tooling',
      'Code: <code>core/onboarding/projection_json.go</code>',
    ]
  },
  html: function (w) {
    return [
      '{',
      '  ' + T('accent', '"status"') + ': ' + T('ready', '"ready"') + ',',
      '  ' + T('accent', '"overall_status"') + ': ' + T('ready', '"ready"') + ',',
      '  ' + T('accent', '"exit_code"') + ': ' + T('ready', '0') + ',',
      '  ' + T('accent', '"checks"') + ': [',
      '    {',
      '      ' + T('accent', '"id"') + ': ' + T('muted', '"docker.binary"') + ',',
      '      ' + T('accent', '"label"') + ': ' + T('muted', '"Docker CLI"') + ',',
      '      ' + T('accent', '"status"') + ': ' + T('ready', '"ready"') + ',',
      '      ' + T('accent', '"message"') + ': ' + T('muted', '"installed"'),
      '    },',
      '    {',
      '      ' + T('accent', '"id"') + ': ' + T('muted', '"docker.daemon"') + ',',
      '      ' + T('accent', '"label"') + ': ' + T('muted', '"Docker daemon"') + ',',
      '      ' + T('accent', '"status"') + ': ' + T('needs-action', '"needs_action"') + ',',
      '      ' + T('accent', '"message"') + ': ' + T('muted', '"not running"') + ',',
      '      ' + T('accent', '"remediation"') + ': ' + T('bright-cyan', '"open -a Docker"'),
      '    }',
      '  ]',
      '}',
    ].join('\n');
  }
};

// 18. GUIDED SHELL — Status utility surface
Screens['guided-status'] = {
  group: 'Guided Shell',
  title: 'Status report (utility surface)',
  annotations: {
    heading: 'Surface: Utility surface \u2014 status report',
    notes: [
      'The result of selecting "Check this project\'s status"',
      'Shows live container status per service',
      'Code: <code>cmd/stage/commands/tui.go \u2192 handleGuidedAction "status"</code>',
    ]
  },
  html: function (w) {
    var lw = w - 4;
    return [
      hdr('Status', lw),
      '  ' + rule(lw),
      '',
      '  ' + T('verdict-ready', 'This project is running at https://demo.test.'),
      '',
      section('Project status', 'neutral', lw),
      '',
      '  ' + T('label', 'Project') + '     demo',
      '  ' + T('label', 'Stack') + '       20i (Apache + PHP)',
      '  ' + T('label', 'Local URL') + '   https://demo.test',
      '',
      '  ' + T('label', 'apache') + '      ' + T('ready', 'running') + '   Up 3 hours',
      '  ' + T('label', 'nginx') + '       ' + T('ready', 'running') + '   Up 3 hours',
      '',
      '  ' + rule(lw),
      '  ' + T('footer', '\u21b5 back to menu  q quit'),
      '',
    ].join('\n');
  }
};

// 19. GUIDED SHELL — Error state
Screens['guided-error'] = {
  group: 'Guided Shell',
  title: 'Action error',
  annotations: {
    heading: 'Surface: Error message after action failure',
    notes: [
      'Shown when a guided action fails (e.g. Docker not running)',
      'Uses RED verdict tone',
      'Message shows the StepError remedy from the lifecycle layer',
      'Code: <code>cmd/stage/commands/tui.go \u2192 renderGuidedActionError</code>',
    ]
  },
  html: function (w) {
    var lw = w - 4;
    return [
      hdr('Starting up', lw),
      '  ' + rule(lw),
      '',
      '  ' + T('verdict-error', "Couldn't start this project."),
      '  ' + T('muted', 'Run stage doctor to check setup, or stage up to retry'),
      '',
      section('What you can do', 'action', lw),
      '',
      '  <span class="tok-cursor-item">' + T('active-marker', '\u25b6') + '  ' + T('focus', 'Run diagnostics') + '                </span>',
      '     Read-only. Check machine and runtime readiness.',
      '',
      '  ' + T('inactive-marker', ' ') + '  Try again                               ',
      '     Retry starting this project.',
      '',
      '',
      '  ' + rule(lw),
      '  ' + T('footer', '\u2191\u2193 navigate  \u21b5 select  q quit'),
      '',
    ].join('\n');
  }
};

// 20. VERSION
Screens['version'] = {
  group: 'CLI Commands',
  title: 'stage version',
  annotations: {
    heading: 'Surface: Version output',
    notes: [
      'Simple single-line version output from <code>stage version</code>',
      'Code: <code>cmd/stage/commands/version.go</code>',
    ]
  },
  html: function (w) {
    return 'StageServe v0.12.0\n';
  }
};

// 21. NARROW TERMINAL
Screens['guided-narrow'] = {
  group: 'Responsive',
  title: 'Narrow terminal (52 cols)',
  annotations: {
    heading: 'Surface: Responsive layout \u2014 narrow terminal',
    notes: [
      'When terminal width < 58, layout adapts:',
      'Key facts stack vertically instead of inline',
      'Confirmation hint text shortens',
      'Code: <code>core/guidance/shell.go \u2192 clampInt</code>',
    ]
  },
  html: function (w) {
    var lw = 48;
    return [
      hdr('Ready', lw),
      '  ' + rule(lw),
      '',
      '  ' + T('verdict-ready', 'This project is ready to run.'),
      '',
      section('Key facts', 'neutral', lw),
      '',
      '  ' + T('label', 'Site name'),
      '    demo',
      '  ' + T('label', 'Local URL'),
      '    https://demo.test',
      '  ' + T('label', 'Web folder'),
      '    public_html',
      '',
      section('What you can do', 'action', lw),
      '',
      '  <span class="tok-cursor-item">' + T('active-marker', '\u25b6') + ' ' + T('focus', 'Run this project') + '</span>',
      '   Start the project and open it',
      '',
      '  ' + T('inactive-marker', ' ') + ' ' + T('label', 'Edit project settings'),
      '   Change name, folder, or address',
      '',
      '  ' + rule(lw),
      '  ' + T('footer', '\u2191\u2193  \u21b5  d  q'),
      '',
    ].join('\n');
  }
};

// 22. NO-COLOR MODE
Screens['guided-nocolor'] = {
  group: 'Responsive',
  title: 'No-colour mode (NO_COLOR)',
  annotations: {
    heading: 'Surface: No-colour fallback',
    notes: [
      'When <code>NO_COLOR</code> is set or terminal doesn\'t support colour',
      'All style tokens become bold-only (no foreground colours)',
      'Code: <code>core/guidance/styles.go \u2192 shellStylesFor(true)</code>',
    ]
  },
  html: function (w) {
    var lw = w - 4;
    return [
      hdr('Ready', lw),
      '  ' + rule(lw),
      '',
      '  This project is ready to run.',
      '',
      section('What you can do', 'neutral', lw),
      '',
      '  \u25b6  Run this project',
      '     Start the project and open it in your browser.',
      '',
      '     Edit project settings',
      '     Change the project name, web folder, or local address.',
      '',
      '  ' + rule(lw),
      '  \u2191\u2193 navigate  \u21b5 select  d details  q quit',
      '',
    ].join('\n');
  }
};

// 23. INIT — Huh form
Screens['init-tui'] = {
  group: 'CLI Commands',
  title: 'stage init (Huh form TUI)',
  annotations: {
    heading: 'Surface: stage init \u2014 interactive Huh form',
    notes: [
      'When TTY is available, stage init shows a Charm Huh form',
      'Fields: project name, web folder, domain suffix',
      'Falls back to plain text projector when no TTY',
      'Code: <code>cmd/stage/commands/init.go \u2192 NewInit</code>',
    ]
  },
  html: function (w) {
    return [
      '',
      '  ' + T('title', 'Set up this project for StageServe'),
      '',
      '  ' + T('label', 'Site name'),
      '  ' + T('focus', '\u2502 demo                                                                \u2502'),
      '  ' + T('muted', '  The name shown in URLs and the StageServe dashboard'),
      '',
      '  ' + T('label', 'Web folder'),
      '  ' + T('muted', '\u2502 public_html                                                         \u2502'),
      '  ' + T('muted', '  Relative path from the project root to the document root'),
      '',
      '  ' + T('label', 'Domain suffix'),
      '  ' + T('muted', '\u2502 test                                                                \u2502'),
      '  ' + T('muted', '  Top-level domain for local development URLs'),
      '',
      '  ' + T('footer', 'tab next  \u21b5 submit  esc cancel'),
      '',
    ].join('\n');
  }
};

// 24. LOGS STREAM
Screens['logs-stream'] = {
  group: 'CLI Commands',
  title: 'stage logs (streaming)',
  annotations: {
    heading: 'Surface: stage logs \u2014 live log stream',
    notes: [
      'Shows streaming container logs for a specific service',
      'Live-updating output with ctrl+c to dismiss',
      'Code: <code>cmd/stage/commands/logs.go \u2192 NewLogs</code>',
    ]
  },
  html: function (w) {
    return [
      T('dim', '\u2500\u2500 apache ' + '\u2500'.repeat(37)),
      '',
      T('muted', '[2026-06-05 14:32:01]') + ' [apache:access] GET / HTTP/1.1 ' + T('ready', '200') + ' 1234',
      T('muted', '[2026-06-05 14:32:01]') + ' [apache:access] GET /clock.html ' + T('ready', '200') + ' 892',
      T('muted', '[2026-06-05 14:32:02]') + ' [apache:access] GET /favicon.ico ' + T('needs-action', '404') + ' 0',
      T('muted', '[2026-06-05 14:32:05]') + ' [apache:error] File not found: /var/www/html/favicon.ico',
      T('muted', '[2026-06-05 14:32:12]') + ' [apache:access] GET / HTTP/1.1 ' + T('ready', '200') + ' 1234',
      '',
      T('dim', '^C to stop \u2022 streaming...'),
    ].join('\n');
  }
};

// Export
function getScreenList() {
  var groups = {};
  var order = [];
  for (var id in Screens) {
    var screen = Screens[id];
    var g = screen.group || 'Other';
    if (!groups[g]) { groups[g] = []; order.push(g); }
    groups[g].push({ id: id, title: screen.title });
  }
  return { groups: groups, order: order };
}
