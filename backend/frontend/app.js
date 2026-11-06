const STATUSES = ["TODO", "DOING", "REVIEW", "DONE"];

let state = {
  tasks: [],
  items: [],
  users: [],
  userFilter: "all",
  activeTaskId: null,
};

async function api(path, options = {}) {
  const res = await fetch(path, {
    headers: { "Content-Type": "application/json", ...(options.headers || {}) },
    ...options,
  });
  if (!res.ok) {
    const msg = await res.text();
    throw new Error(msg || res.statusText);
  }
  if (res.status === 204) return null;
  return res.json();
}

// Tabs
document.querySelectorAll(".tab").forEach((btn) => {
  btn.addEventListener("click", () => {
    document.querySelectorAll(".tab").forEach((b) => b.classList.remove("active"));
    document.querySelectorAll(".view").forEach((v) => v.classList.remove("active"));
    btn.classList.add("active");
    document.getElementById(`view-${btn.dataset.tab}`).classList.add("active");
  });
});

// Auth (/api/me)
async function loadMe() {
  const badge = document.getElementById("me-badge");
  try {
    const me = await api("/api/me");
    badge.textContent = `user: ${me.username} (${me.role || "sin rol"})`;
  } catch {
    badge.textContent = "auth: local / sin sesión CF";
  }
}

// Tasks / Kanban
async function loadTasks() {
  state.tasks = await api("/api/tasks");
  renderBoard();
}

function renderBoard() {
  STATUSES.forEach((st) => {
    document.getElementById(`col-${st}`).innerHTML = "";
  });

  state.tasks.forEach((t) => {
    const col = document.getElementById(`col-${t.status}`);
    if (!col) return;

    const idx = STATUSES.indexOf(t.status);
    const prevStatus = idx > 0 ? STATUSES[idx - 1] : null;
    const nextStatus = idx < STATUSES.length - 1 ? STATUSES[idx + 1] : null;

    const card = document.createElement("div");
    card.className = "card";
    card.innerHTML = `
      <div class="card-header">
        <span class="card-title" data-open="${t.id}">#${t.id} ${escapeHTML(t.title)}</span>
        <span class="prio-${t.priority}">[${t.priority}]</span>
      </div>
      <div class="card-desc">${escapeHTML(t.description)}</div>
      <div class="card-controls">
        <div>
          ${prevStatus ? `<button data-move="${t.id}" data-to="${prevStatus}">&lt; ${prevStatus}</button>` : ""}
          ${nextStatus ? `<button data-move="${t.id}" data-to="${nextStatus}">${nextStatus} &gt;</button>` : ""}
        </div>
        <button data-open="${t.id}">detalle</button>
      </div>
    `;
    col.appendChild(card);
  });
}

document.querySelector(".board").addEventListener("click", async (e) => {
  const moveBtn = e.target.closest("[data-move]");
  if (moveBtn) {
    const id = moveBtn.dataset.move;
    const status = moveBtn.dataset.to;
    await api(`/api/tasks/${id}/status`, {
      method: "PATCH",
      body: JSON.stringify({ status }),
    });
    await loadTasks();
    return;
  }

  const openEl = e.target.closest("[data-open]");
  if (openEl) {
    openTaskDialog(Number(openEl.dataset.open));
  }
});

document.getElementById("form-new-task").addEventListener("submit", async (e) => {
  e.preventDefault();
  const fd = new FormData(e.target);
  await api("/api/tasks", {
    method: "POST",
    body: JSON.stringify({
      title: fd.get("title"),
      description: fd.get("description"),
    }),
  });
  e.target.reset();
  await loadTasks();
});

// Task Modal (Edit + Assignees + Items)
const dialog = document.getElementById("task-dialog");
document.getElementById("btn-close-dialog").addEventListener("click", () => dialog.close());

async function openTaskDialog(taskId) {
  state.activeTaskId = taskId;
  const task = await api(`/api/tasks/${taskId}`);
  const form = document.getElementById("form-edit-task");
  document.getElementById("dialog-task-id").textContent = `Tarea #${task.id} (${task.status})`;
  form.elements.id.value = task.id;
  form.elements.title.value = task.title;
  form.elements.description.value = task.description;
  form.elements.priority.value = task.priority;

  await refreshTaskRelations(taskId);
  dialog.showModal();
}

async function refreshTaskRelations(taskId) {
  const [assignees, taskItems] = await Promise.all([
    api(`/api/tasks/${taskId}/assignees`),
    api(`/api/tasks/${taskId}/items`),
  ]);

  const assigneesList = document.getElementById("task-assignees-list");
  assigneesList.innerHTML = assignees
    .map(
      (u) => `<li>
        <span>${escapeHTML(u.name)} (@${escapeHTML(u.username)})</span>
        <button type="button" data-unassign-user="${u.id}">x</button>
      </li>`
    )
    .join("");

  const itemsList = document.getElementById("task-items-list");
  itemsList.innerHTML = taskItems
    .map(
      (i) => `<li>
        <span>#${i.id} [${escapeHTML(i.type)}] ${escapeHTML(i.description || "")}</span>
        <button type="button" data-unassign-item="${i.id}">x</button>
      </li>`
    )
    .join("");

  const userSelect = document.getElementById("select-assign-user");
  userSelect.innerHTML = state.users
    .map((u) => `<option value="${u.id}">${escapeHTML(u.name)} (@${escapeHTML(u.username)})</option>`)
    .join("");

  const itemSelect = document.getElementById("select-assign-item");
  itemSelect.innerHTML = state.items
    .map((i) => `<option value="${i.id}">#${i.id} ${escapeHTML(i.type)} - ${escapeHTML(i.description || "")}</option>`)
    .join("");
}

document.getElementById("form-edit-task").addEventListener("submit", async (e) => {
  e.preventDefault();
  const fd = new FormData(e.target);
  const id = fd.get("id");
  await api(`/api/tasks/${id}`, {
    method: "PUT",
    body: JSON.stringify({
      title: fd.get("title"),
      description: fd.get("description"),
      priority: fd.get("priority"),
    }),
  });
  dialog.close();
  await loadTasks();
});

document.getElementById("btn-delete-task").addEventListener("click", async () => {
  if (!state.activeTaskId) return;
  await api(`/api/tasks/${state.activeTaskId}`, { method: "DELETE" });
  dialog.close();
  await loadTasks();
});

document.getElementById("form-assign-user").addEventListener("submit", async (e) => {
  e.preventDefault();
  const userId = Number(document.getElementById("select-assign-user").value);
  if (!userId || !state.activeTaskId) return;
  await api("/api/tasks/assign", {
    method: "POST",
    body: JSON.stringify({ user_id: userId, task_id: state.activeTaskId }),
  });
  await refreshTaskRelations(state.activeTaskId);
});

document.getElementById("task-assignees-list").addEventListener("click", async (e) => {
  const btn = e.target.closest("[data-unassign-user]");
  if (!btn) return;
  await api(`/api/tasks/${state.activeTaskId}/assignees/${btn.dataset.unassignUser}`, {
    method: "DELETE",
  });
  await refreshTaskRelations(state.activeTaskId);
});

document.getElementById("form-assign-item").addEventListener("submit", async (e) => {
  e.preventDefault();
  const itemId = Number(document.getElementById("select-assign-item").value);
  if (!itemId || !state.activeTaskId) return;
  await api("/api/items/assign", {
    method: "POST",
    body: JSON.stringify({ task_id: state.activeTaskId, item_id: itemId }),
  });
  await refreshTaskRelations(state.activeTaskId);
});

document.getElementById("task-items-list").addEventListener("click", async (e) => {
  const btn = e.target.closest("[data-unassign-item]");
  if (!btn) return;
  await api(`/api/tasks/${state.activeTaskId}/items/${btn.dataset.unassignItem}`, {
    method: "DELETE",
  });
  await refreshTaskRelations(state.activeTaskId);
});

// Inventory Items
async function loadItems() {
  state.items = await api("/api/items");
  const tbody = document.getElementById("items-tbody");
  tbody.innerHTML = state.items
    .map(
      (i) => `<tr>
        <td>#${i.id}</td>
        <td>${escapeHTML(i.type)}</td>
        <td>${escapeHTML(i.description || "-")}</td>
        <td><button data-del-item="${i.id}" class="danger">borrar</button></td>
      </tr>`
    )
    .join("");
}

document.getElementById("form-new-item").addEventListener("submit", async (e) => {
  e.preventDefault();
  const fd = new FormData(e.target);
  const desc = fd.get("description")?.trim();
  await api("/api/items", {
    method: "POST",
    body: JSON.stringify({
      type: fd.get("type"),
      description: desc ? desc : null,
    }),
  });
  e.target.reset();
  await loadItems();
});

document.getElementById("items-tbody").addEventListener("click", async (e) => {
  const btn = e.target.closest("[data-del-item]");
  if (!btn) return;
  await api(`/api/items/${btn.dataset.delItem}`, { method: "DELETE" });
  await loadItems();
});

// Users
async function loadUsers() {
  let endpoint = "/api/users";
  if (state.userFilter === "with-tasks") endpoint = "/api/users/with-tasks";
  if (state.userFilter === "without-tasks") endpoint = "/api/users/without-tasks";

  const list = await api(endpoint);
  if (state.userFilter === "all") state.users = list;

  const tbody = document.getElementById("users-tbody");
  tbody.innerHTML = list
    .map(
      (u) => `<tr>
        <td>#${u.id}</td>
        <td>${escapeHTML(u.username)}</td>
        <td>${escapeHTML(u.name)}</td>
        <td>${escapeHTML(u.role || "-")}</td>
      </tr>`
    )
    .join("");
}

document.querySelectorAll("[data-user-filter]").forEach((btn) => {
  btn.addEventListener("click", () => {
    document.querySelectorAll("[data-user-filter]").forEach((b) => b.classList.remove("active"));
    btn.classList.add("active");
    state.userFilter = btn.dataset.userFilter;
    loadUsers();
  });
});

document.getElementById("form-new-user").addEventListener("submit", async (e) => {
  e.preventDefault();
  const fd = new FormData(e.target);
  const role = fd.get("role")?.trim();
  await api("/api/users", {
    method: "POST",
    body: JSON.stringify({
      username: fd.get("username"),
      name: fd.get("name"),
      role: role ? role : null,
    }),
  });
  e.target.reset();
  await loadUsers();
});

function escapeHTML(str) {
  const d = document.createElement("div");
  d.textContent = str;
  return d.innerHTML;
}

// Init
Promise.all([loadMe(), loadUsers(), loadItems(), loadTasks()]);
