const state = {
  health: null,
  logs: []
};

const formatTime = (value) => {
  if (!value || value.startsWith("0001-")) return "-";
  return new Date(value).toLocaleTimeString();
};

const text = (value) => value === "" || value === undefined || value === null ? "-" : value;
const escapeHTML = (value) => String(value)
  .replaceAll("&", "&amp;")
  .replaceAll("<", "&lt;")
  .replaceAll(">", "&gt;")
  .replaceAll('"', "&quot;")
  .replaceAll("'", "&#039;");

async function fetchJSON(path) {
  const response = await fetch(path, { cache: "no-store" });
  if (!response.ok && response.status !== 503) {
    throw new Error(`${path} returned ${response.status}`);
  }
  return response.json();
}

async function refresh() {
  try {
    const [health, logs] = await Promise.all([
      fetchJSON("/health"),
      fetchJSON("/logs")
    ]);
    state.health = health;
    state.logs = logs.logs || [];
    render();
  } catch (error) {
    document.getElementById("overallStatus").textContent = "Offline";
    document.getElementById("overallStatus").className = "status-pill error";
  }
}

function render() {
  const health = state.health;
  const overall = document.getElementById("overallStatus");
  overall.textContent = health.status;
  overall.className = `status-pill ${health.status === "error" ? "error" : ""}`;

  document.getElementById("watchersOk").textContent = health.watchersOk;
  document.getElementById("errors").textContent = health.errors;
  document.getElementById("checkedAt").textContent = formatTime(health.checkedAt);
  document.getElementById("watcherCount").textContent = `${health.watchers.length} configured`;

  const watchers = document.getElementById("watchers");
  const orderedWatchers = [...health.watchers].sort((left, right) => {
    const feature = left.feature.localeCompare(right.feature);
    if (feature !== 0) return feature;
    return left.name.localeCompare(right.name);
  });

  watchers.innerHTML = orderedWatchers.map((watcher) => `
    <article class="watcher">
      <div class="watcher-head">
        <div>
          <h3>${escapeHTML(watcher.name)}</h3>
          <div class="meta">${escapeHTML(watcher.feature)}</div>
        </div>
        <span class="badge ${watcher.status === "error" ? "error" : ""}">${escapeHTML(watcher.status)}</span>
      </div>
      <div class="grid">
        <span>Alive</span><strong>${watcher.alive}</strong>
        <span>Samba</span><strong>${watcher.sambaEnabled ? (watcher.sambaConnected ? "connected" : "disconnected") : "disabled"}</strong>
        <span>Last scan</span><strong>${formatTime(watcher.lastScanAt)}</strong>
        <span>Heartbeat</span><strong>${formatTime(watcher.lastHeartbeatAt)}</strong>
        <span>Path</span><strong class="path">${escapeHTML(text(watcher.path))}</strong>
        ${watcher.lastError ? `<span>Error</span><strong class="error-text">${escapeHTML(watcher.lastError)}</strong>` : ""}
      </div>
    </article>
  `).join("");

  document.getElementById("logCount").textContent = `${state.logs.length} lines`;
  const logs = document.getElementById("logs");
  logs.innerHTML = state.logs.map((entry) => `
    <div class="log-line">
      <span class="log-time">${formatTime(entry.time)}</span>
      <span class="log-level ${escapeHTML(entry.level)}">${escapeHTML(entry.level)}</span>
      <span class="log-message">${escapeHTML(entry.message)}${formatAttrs(entry.attrs)}</span>
    </div>
  `).join("");
  logs.scrollTop = logs.scrollHeight;
}

function formatAttrs(attrs) {
  if (!attrs || Object.keys(attrs).length === 0) return "";
  return " " + Object.entries(attrs)
    .map(([key, value]) => `${escapeHTML(key)}=${escapeHTML(value)}`)
    .join(" ");
}

refresh();
setInterval(refresh, 2000);
