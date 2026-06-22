const $ = (sel) => document.querySelector(sel);

let catalog = [];
let categories = [];
let unavailableState = { byId: {}, byType: {} };
let activeId = null;
const collapsed = new Set();
let paramNotes = {};
const collapsedParamPaths = new Set();

async function apiFetch(url, options = {}) {
  const res = await fetch(url, options);
  if (res.status === 401) {
    window.location.href = "/login.html";
    throw new Error("未登录");
  }
  return res;
}

async function initAuth() {
  const res = await fetch("/api/auth/me");
  const data = await res.json();
  const logoutBtn = $("#logout-btn");
  if (data.authEnabled) {
    logoutBtn.hidden = false;
    if (!data.authenticated) {
      window.location.href = "/login.html";
      return false;
    }
  }
  return true;
}

async function logout() {
  await fetch("/api/auth/logout", { method: "POST" });
  window.location.href = "/login.html";
}

async function loadConfig() {
  const res = await apiFetch("/api/config");
  const cfg = await res.json();
  $("#config-badge").textContent =
    `env=${cfg.env} | region=${cfg.region} | key=${cfg.app_key}`;
}

async function loadCatalog() {
  const res = await apiFetch("/api/catalog");
  const data = await res.json();
  categories = data.categories || [];
  catalog = data.apis || data;
  unavailableState = data.unavailable || { byId: {}, byType: {} };
  renderCatalog(filterCatalog($("#search").value));
  updateMarkUnavailableButton();
  if (activeId) {
    const api = catalog.find((a) => a.id === activeId);
    if (api) updateApiDesc(api);
  }
}

function isApiUnavailable(api) {
  if (!api) return isTypeUnavailable(api?.type);
  if (api.unavailable) return true;
  return isTypeUnavailable(api.type);
}

function isTypeUnavailable(type) {
  const t = (type || "").trim();
  if (!t) return false;
  if (unavailableState.byType && unavailableState.byType[t]) return true;
  const api = catalog.find((a) => a.type === t);
  if (api && unavailableState.byId && unavailableState.byId[api.id]) return true;
  return false;
}

function unavailableNote(api) {
  if (!api) {
    const t = ($("#api-type")?.value || "").trim();
    return unavailableState.byType?.[t]?.note || "";
  }
  const byId = unavailableState.byId?.[api.id]?.note;
  if (byId) return byId;
  const byType = unavailableState.byType?.[api.type]?.note;
  if (byType) return byType;
  return api.unavailableNote || "";
}

function valueType(v) {
  if (v === null) return "null";
  if (Array.isArray(v)) return "array";
  return typeof v;
}

function previewValue(v) {
  if (v === null) return "null";
  if (typeof v === "string") {
    if (v.length > 48) return v.slice(0, 45) + "...";
    return v;
  }
  if (typeof v === "object") {
    return Array.isArray(v) ? `[${v.length}]` : "{...}";
  }
  return String(v);
}

function getNoteAtPath(notes, path) {
  if (!notes || !path) return "";
  const parts = path.split(".");
  let cur = notes;
  for (const part of parts) {
    if (cur == null || typeof cur !== "object") return "";
    cur = cur[part];
  }
  return typeof cur === "string" ? cur : "";
}

function setNoteAtPath(notes, path, value) {
  const parts = path.split(".");
  let cur = notes;
  for (let i = 0; i < parts.length - 1; i++) {
    const p = parts[i];
    if (cur[p] == null || typeof cur[p] !== "object" || Array.isArray(cur[p])) {
      cur[p] = {};
    }
    cur = cur[p];
  }
  const last = parts[parts.length - 1];
  if (!value.trim()) {
    delete cur[last];
  } else {
    cur[last] = value.trim();
  }
}

function pruneEmptyNotes(obj) {
  if (!obj || typeof obj !== "object" || Array.isArray(obj)) return obj;
  const out = {};
  for (const [k, v] of Object.entries(obj)) {
    if (typeof v === "string") {
      if (v.trim()) out[k] = v.trim();
    } else if (v && typeof v === "object") {
      const child = pruneEmptyNotes(v);
      if (child && Object.keys(child).length) out[k] = child;
    }
  }
  return out;
}

function pathSegment(path) {
  const m = path.match(/(\[[0-9]+\]|[^.[\]]+)$/);
  return m ? m[1] : path;
}

function parentPath(path) {
  if (!path) return "";
  const m = path.match(/^(.*)(?:\.[^.[\]]+|\[[0-9]+\])$/);
  return m ? m[1] : "";
}

function hasChildNodes(value) {
  const t = valueType(value);
  if (t === "array") return Array.isArray(value) && value.length > 0;
  if (t === "object" && value !== null) return Object.keys(value).length > 0;
  return false;
}

function collectParamRows(value, path, depth, rows) {
  const t = valueType(value);
  const hasChildren = hasChildNodes(value);
  rows.push({
    path,
    segment: pathSegment(path),
    type: t,
    preview: previewValue(value),
    depth,
    hasChildren,
  });

  if (t === "object" && value !== null && !Array.isArray(value)) {
    for (const key of Object.keys(value)) {
      const childPath = path ? `${path}.${key}` : key;
      collectParamRows(value[key], childPath, depth + 1, rows);
    }
  } else if (t === "array") {
    value.forEach((item, idx) => {
      collectParamRows(item, `${path}[${idx}]`, depth + 1, rows);
    });
  }
}

function isPathVisible(path) {
  const parent = parentPath(path);
  if (!parent) return true;
  if (collapsedParamPaths.has(parent)) return false;
  return isPathVisible(parent);
}

function buildParamRows(parsed) {
  const rows = [];
  for (const key of Object.keys(parsed)) {
    collectParamRows(parsed[key], key, 0, rows);
  }
  return rows;
}

function applyDefaultCollapse(rows) {
  collapsedParamPaths.clear();
  for (const row of rows) {
    if (row.hasChildren && row.depth >= 1) {
      collapsedParamPaths.add(row.path);
    }
  }
}

function parseRequestBodyObject(bodyText) {
  if (!bodyText) return { error: "请先填写请求 Body" };
  try {
    const parsed = JSON.parse(bodyText);
    if (parsed === null || typeof parsed !== "object" || Array.isArray(parsed)) {
      return { error: "Body 需为 JSON 对象才能展开参数" };
    }
    return { parsed };
  } catch {
    return { error: "Body 不是合法 JSON，无法生成参数表" };
  }
}

function createExpandControl(row) {
  if (!row.hasChildren) {
    const spacer = document.createElement("span");
    spacer.className = "param-expand-placeholder";
    return spacer;
  }
  const expanded = !collapsedParamPaths.has(row.path);
  const btn = document.createElement("button");
  btn.type = "button";
  btn.className = "param-expand-btn" + (expanded ? " is-expanded" : "");
  btn.textContent = expanded ? "▼" : "▶";
  btn.title = expanded ? "收起子参数" : "展开子参数";
  btn.addEventListener("click", (e) => {
    e.stopPropagation();
    if (collapsedParamPaths.has(row.path)) collapsedParamPaths.delete(row.path);
    else collapsedParamPaths.add(row.path);
    renderParamNotesTable(false);
  });
  return btn;
}

function renderParamNotesTable(resetCollapse = true) {
  const tbody = $("#param-notes-body");
  if (!tbody) return;

  const bodyText = $("#request-body").value.trim();
  const { parsed, error } = parseRequestBodyObject(bodyText);
  if (error) {
    tbody.innerHTML = `<tr><td colspan="4" class="empty-hint">${error}</td></tr>`;
    return;
  }

  const rows = buildParamRows(parsed);
  if (resetCollapse) applyDefaultCollapse(rows);

  tbody.innerHTML = "";
  for (const row of rows) {
    if (!isPathVisible(row.path)) continue;

    const tr = document.createElement("tr");
    tr.className = "param-row";
    tr.dataset.depth = String(row.depth);
    if (row.hasChildren) tr.classList.add("param-row--branch");
    if (collapsedParamPaths.has(row.path)) tr.classList.add("param-row--collapsed");

    const pathTd = document.createElement("td");
    pathTd.className = "col-path";

    const cell = document.createElement("div");
    cell.className = "param-path-cell";
    cell.style.paddingLeft = `${8 + row.depth * 18}px`;

    cell.appendChild(createExpandControl(row));

    const segment = document.createElement("span");
    segment.className = "param-segment";
    segment.textContent = row.segment;
    segment.title = row.path;
    cell.appendChild(segment);

    pathTd.appendChild(cell);

    const typeTd = document.createElement("td");
    typeTd.className = "col-type";
    typeTd.textContent = row.type;

    const previewTd = document.createElement("td");
    previewTd.className = "col-preview";
    previewTd.textContent = row.preview;

    const noteTd = document.createElement("td");
    const input = document.createElement("input");
    input.type = "text";
    input.className = "param-note-input";
    input.placeholder = "填写参数说明…";
    input.value = getNoteAtPath(paramNotes, row.path);
    input.dataset.path = row.path;
    input.addEventListener("input", (e) => {
      setNoteAtPath(paramNotes, e.target.dataset.path, e.target.value);
    });
    noteTd.appendChild(input);

    tr.appendChild(pathTd);
    tr.appendChild(typeTd);
    tr.appendChild(previewTd);
    tr.appendChild(noteTd);
    tbody.appendChild(tr);
  }
}

function showSampleStatus(message, ok) {
  const el = $("#sample-save-status");
  if (!el) return;
  el.hidden = false;
  el.textContent = message;
  el.className = "sample-save-status " + (ok ? "ok" : "error");
  if (ok) {
    setTimeout(() => {
      el.hidden = true;
    }, 3000);
  }
}

async function saveSample() {
  const { api, type } = currentApiContext();
  if (!type) {
    alert("请填写 API Type 或选择接口");
    return false;
  }

  const bodyText = $("#request-body").value.trim();
  let body = {};
  if (bodyText) {
    try {
      body = JSON.parse(bodyText);
    } catch {
      alert("请求 Body 不是合法的 JSON");
      return false;
    }
  }

  const res = await apiFetch("/api/samples", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      id: api?.id || "",
      type,
      body,
      paramNotes: pruneEmptyNotes(paramNotes),
    }),
  });
  const data = await res.json();
  if (!res.ok || data.error) {
    showSampleStatus(data.error || "保存失败", false);
    return false;
  }
  if (data.apis) {
    catalog = data.apis;
    renderCatalog(filterCatalog($("#search").value));
  }
  showSampleStatus("已保存到 api-samples.json", true);
  return true;
}

async function resetSample() {
  const { api, type } = currentApiContext();
  if (!type) {
    alert("请填写 API Type 或选择接口");
    return;
  }
  if (!confirm("确定恢复该接口的内置默认示例？将清除 api-samples.json 中的自定义 Body 与备注。")) {
    return;
  }

  const res = await apiFetch("/api/samples", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      id: api?.id || "",
      type,
      clear: true,
    }),
  });
  const data = await res.json();
  if (!res.ok || data.error) {
    showSampleStatus(data.error || "恢复失败", false);
    return;
  }
  if (data.apis) catalog = data.apis;
  await loadCatalog();
  const fresh = api?.id ? catalog.find((a) => a.id === api.id) : catalog.find((a) => a.type === type);
  if (fresh) selectApi(fresh);
  showSampleStatus("已恢复内置默认", true);
}

async function saveUnavailableMark(api, type, unavailable, note) {
  const body = {
    unavailable,
    note: (note || "").trim(),
    type: type || api?.type || "",
  };
  if (api?.id) body.id = api.id;

  if (!body.type && !body.id) {
    alert("请先选择接口或填写 API Type");
    return false;
  }

  const res = await apiFetch("/api/unavailable", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  const data = await res.json();
  if (!res.ok || data.error) {
    alert(data.error || "保存失败");
    return false;
  }
  unavailableState = data.unavailable || unavailableState;
  await loadCatalog();
  return true;
}

function renderCatalog(items) {
  const nav = $("#api-list");
  nav.innerHTML = "";

  const grouped = {};
  for (const item of items) {
    (grouped[item.category] ||= []).push(item);
  }

  const order = categories.length
    ? categories.filter((c) => grouped[c])
    : Object.keys(grouped).sort();

  for (const category of order) {
    const apis = grouped[category];
    if (!apis?.length) continue;

    const section = document.createElement("div");
    section.className = "category-section";

    const isCollapsed = collapsed.has(category);
    const header = document.createElement("button");
    header.className = "category-header" + (isCollapsed ? " collapsed" : "");
    header.innerHTML =
      `<span class="category-arrow">${isCollapsed ? "▶" : "▼"}</span>` +
      `<span class="category-label">${category}</span>` +
      `<span class="category-count">${apis.length}</span>`;
    header.addEventListener("click", () => {
      if (collapsed.has(category)) collapsed.delete(category);
      else collapsed.add(category);
      renderCatalog(filterCatalog($("#search").value));
    });
    section.appendChild(header);

    if (!isCollapsed) {
      const list = document.createElement("div");
      list.className = "category-items";
      for (const api of apis) {
        const row = document.createElement("div");
        row.className = "api-item" + (api.id === activeId ? " active" : "");

        const star = document.createElement("button");
        star.type = "button";
        star.className =
          "api-mark-star" + (isApiUnavailable(api) ? " is-marked" : "");
        star.textContent = "★";
        star.title = isApiUnavailable(api)
          ? "取消不可用标记"
          : "标记为不可用";
        star.addEventListener("click", async (e) => {
          e.stopPropagation();
          const marked = isApiUnavailable(api);
          const note = marked
            ? ""
            : ($("#unavailable-note")?.value || "").trim();
          await saveUnavailableMark(api, api.type, !marked, note);
        });

        const btn = document.createElement("button");
        btn.type = "button";
        btn.className = "api-item-body";
        const marked = isApiUnavailable(api);
        const typeClass = marked ? "api-type unavailable" : "api-type";
        btn.innerHTML =
          `<span class="api-name">${api.name}</span>` +
          `<span class="${typeClass}">${api.type}</span>`;
        btn.addEventListener("click", () => selectApi(api));

        row.appendChild(star);
        row.appendChild(btn);
        list.appendChild(row);
      }
      section.appendChild(list);
    }

    nav.appendChild(section);
  }

  if (!items.length) {
    nav.innerHTML = '<div class="empty-hint">无匹配接口</div>';
  }
}

function updateApiDesc(api) {
  const marked = isApiUnavailable(api);
  const note = unavailableNote(api);
  const unavailableTip = marked
    ? `<span class="api-unavailable-tip">该接口在 Playground 中无法正常调试${note ? `：${note}` : ""}</span>`
    : "";
  $("#api-desc").innerHTML = api?.description
    ? `<strong>${api.name}</strong>${unavailableTip}<span>${api.description}</span>`
    : unavailableTip;
}

function selectApi(api) {
  activeId = api.id;
  $("#api-type").value = api.type;
  updateApiDesc(api);
  const noteInput = $("#unavailable-note");
  if (noteInput) {
    noteInput.value = isApiUnavailable(api) ? unavailableNote(api) : "";
  }
  paramNotes = api.paramNotes ? JSON.parse(JSON.stringify(api.paramNotes)) : {};
  try {
    $("#request-body").value = JSON.stringify(JSON.parse(api.sampleBody), null, 2);
  } catch {
    $("#request-body").value = api.sampleBody || "{}";
  }
  renderParamNotesTable(true);
  renderCatalog(filterCatalog($("#search").value));
  updateMarkUnavailableButton();
}

function currentApiContext() {
  const type = $("#api-type").value.trim();
  const api = activeId
    ? catalog.find((a) => a.id === activeId)
    : catalog.find((a) => a.type === type);
  return { api, type };
}

function updateMarkUnavailableButton() {
  const btn = $("#mark-unavailable-btn");
  const noteInput = $("#unavailable-note");
  if (!btn) return;
  const { api, type } = currentApiContext();
  const marked = api ? isApiUnavailable(api) : isTypeUnavailable(type);
  btn.textContent = marked ? "取消不可用标记" : "标记不可用";
  btn.classList.toggle("is-marked", marked);
  btn.disabled = !type;
  if (noteInput) {
    noteInput.value = marked ? unavailableNote(api || { type }) : noteInput.value;
    if (!marked && noteInput.value && api && unavailableNote(api) === "") {
      // 保留用户正在输入的备注
    }
  }
}

async function toggleUnavailableMark() {
  const { api, type } = currentApiContext();
  if (!type) return;

  const marked = api ? isApiUnavailable(api) : isTypeUnavailable(type);
  const noteInput = $("#unavailable-note");
  const note = marked ? "" : (noteInput?.value || "").trim();

  await saveUnavailableMark(api, type, !marked, note);
}

function filterCatalog(query) {
  const q = query.trim().toLowerCase();
  if (!q) return catalog;
  const filtered = catalog.filter(
    (a) =>
      a.name.toLowerCase().includes(q) ||
      a.type.toLowerCase().includes(q) ||
      a.category.toLowerCase().includes(q) ||
      (a.description || "").toLowerCase().includes(q)
  );
  if (q) {
    for (const item of filtered) collapsed.delete(item.category);
  }
  return filtered;
}

async function sendRequest() {
  const type = $("#api-type").value.trim();
  const bodyText = $("#request-body").value.trim();
  const btn = $("#send-btn");
  const respEl = $("#response");
  const metaEl = $("#meta");

  if (!type) {
    alert("请填写 API Type");
    return;
  }

  let body = {};
  if (bodyText) {
    try {
      body = JSON.parse(bodyText);
    } catch {
      alert("请求 Body 不是合法的 JSON");
      return;
    }
  }

  btn.disabled = true;
  btn.textContent = "请求中...";
  respEl.textContent = "请求中...";
  respEl.className = "response";
  metaEl.textContent = "";

  try {
    const res = await apiFetch("/api/invoke", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ type, body }),
    });
    const data = await res.json();

    const parts = [];
    if (data.statusCode) parts.push(`${data.statusCode}`);
    if (data.durationMs != null) parts.push(`${data.durationMs}ms`);
    if (data.type) parts.push(data.type);
    metaEl.textContent = parts.join(" · ");

    respEl.textContent = JSON.stringify(data, null, 2);
    respEl.className = "response " + (data.ok ? "ok" : "error");
  } catch (err) {
    respEl.textContent = "网络错误: " + err.message;
    respEl.className = "response error";
  } finally {
    btn.disabled = false;
    btn.textContent = "发送请求";
  }
}

function formatJSON() {
  const ta = $("#request-body");
  try {
    ta.value = JSON.stringify(JSON.parse(ta.value), null, 2);
  } catch {
    alert("JSON 格式不正确");
  }
}

$("#send-btn").addEventListener("click", sendRequest);
$("#format-btn").addEventListener("click", formatJSON);
$("#clear-btn").addEventListener("click", () => {
  $("#request-body").value = "{}";
  $("#response").textContent = "等待请求...";
  $("#response").className = "response";
  $("#meta").textContent = "";
  $("#api-desc").innerHTML = "";
  activeId = null;
  paramNotes = {};
  collapsedParamPaths.clear();
  renderParamNotesTable(true);
  const noteInput = $("#unavailable-note");
  if (noteInput) noteInput.value = "";
  updateMarkUnavailableButton();
});
$("#save-sample-btn").addEventListener("click", saveSample);
$("#refresh-params-btn").addEventListener("click", () => renderParamNotesTable(true));
$("#reset-sample-btn").addEventListener("click", resetSample);
$("#mark-unavailable-btn").addEventListener("click", toggleUnavailableMark);
$("#logout-btn").addEventListener("click", logout);
$("#api-type").addEventListener("input", updateMarkUnavailableButton);
$("#search").addEventListener("input", (e) => {
  renderCatalog(filterCatalog(e.target.value));
});

document.addEventListener("keydown", (e) => {
  if ((e.ctrlKey || e.metaKey) && e.key === "Enter") {
    sendRequest();
  }
});

(async () => {
  if (!(await initAuth())) return;
  loadConfig();
  loadCatalog();
})();
