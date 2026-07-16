const el = (id) => document.getElementById(id);

function logLine(text) {
  const box = el("logBox");
  const line = document.createElement("div");
  const time = new Date().toLocaleTimeString("pl-PL");
  line.textContent = `[${time}] ${text}`;
  box.prepend(line);
  while (box.children.length > 50) box.removeChild(box.lastChild);
}

function statusClass(status) {
  return status.toLowerCase().replace(/\s+/g, "");
}

function formatUptime(seconds) {
  const h = Math.floor(seconds / 3600);
  const m = Math.floor((seconds % 3600) / 60);
  const s = seconds % 60;
  return `${h}h ${m}m ${s}s`;
}

async function fetchStatus() {
  try {
    const res = await fetch("/api/status");
    const data = await res.json();
    el("statRequests").textContent = data.requestCount;
    el("statItems").textContent = data.itemCount;
    el("statGoroutines").textContent = data.goroutines;
    el("statMem").textContent = data.memAllocMB;
    el("uptimeBadge").textContent = `uptime: ${formatUptime(data.uptimeSeconds)}`;
    el("goVersion").textContent = data.goVersion;
    el("healthDot").className = "dot ok";
  } catch (err) {
    el("healthDot").className = "dot bad";
  }
}

async function fetchItems() {
  const res = await fetch("/api/items");
  const items = await res.json();
  const body = el("itemsBody");
  body.innerHTML = "";

  if (!items || items.length === 0) {
    body.innerHTML = `<tr><td colspan="5" class="empty">Brak elementów — dodaj pierwszy powyżej.</td></tr>`;
    return;
  }

  items
    .sort((a, b) => a.id.localeCompare(b.id, undefined, { numeric: true }))
    .forEach((item) => {
      const tr = document.createElement("tr");
      const created = new Date(item.createdAt).toLocaleString("pl-PL");
      tr.innerHTML = `
        <td>#${item.id}</td>
        <td>${escapeHtml(item.name)}</td>
        <td><span class="status-pill ${statusClass(item.status)}">${escapeHtml(item.status)}</span></td>
        <td>${created}</td>
        <td class="row-actions">
          <button class="icon-btn" data-action="cycle" data-id="${item.id}">⟳</button>
          <button class="icon-btn" data-action="delete" data-id="${item.id}">✕</button>
        </td>
      `;
      body.appendChild(tr);
    });
}

function escapeHtml(str) {
  const div = document.createElement("div");
  div.textContent = str;
  return div.innerHTML;
}

const STATUS_CYCLE = ["nowy", "w toku", "gotowe"];

async function cycleStatus(id, currentEl) {
  const currentText = currentEl.textContent.trim();
  const idx = STATUS_CYCLE.indexOf(currentText);
  const next = STATUS_CYCLE[(idx + 1) % STATUS_CYCLE.length];
  await fetch(`/api/items/${id}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ status: next }),
  });
  logLine(`Zmieniono status elementu #${id} -> ${next}`);
  await fetchItems();
  await fetchStatus();
}

async function deleteItem(id) {
  await fetch(`/api/items/${id}`, { method: "DELETE" });
  logLine(`Usunięto element #${id}`);
  await fetchItems();
  await fetchStatus();
}

el("itemsBody").addEventListener("click", (e) => {
  const btn = e.target.closest("button[data-action]");
  if (!btn) return;
  const id = btn.dataset.id;
  if (btn.dataset.action === "delete") deleteItem(id);
  if (btn.dataset.action === "cycle") {
    const pill = btn.closest("tr").querySelector(".status-pill");
    cycleStatus(id, pill);
  }
});

el("createForm").addEventListener("submit", async (e) => {
  e.preventDefault();
  const name = el("newName").value.trim();
  const status = el("newStatus").value;
  if (!name) return;

  const res = await fetch("/api/items", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ name, status }),
  });
  const item = await res.json();
  logLine(`Dodano element #${item.id}: "${name}"`);
  el("newName").value = "";
  await fetchItems();
  await fetchStatus();
});

el("clearLog").addEventListener("click", () => {
  el("logBox").innerHTML = "";
});

async function refreshAll() {
  await fetchStatus();
  await fetchItems();
}

refreshAll();
logLine("Panel załadowany.");
setInterval(fetchStatus, 3000);
