const $ = (sel) => document.querySelector(sel);

let catalog = [];
let categories = [];
let unavailableState = { byId: {}, byType: {} };
let activeId = null;
const collapsed = new Set();

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
  try {
    $("#request-body").value = JSON.stringify(JSON.parse(api.sampleBody), null, 2);
  } catch {
    $("#request-body").value = api.sampleBody || "{}";
  }
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
  const noteInput = $("#unavailable-note");
  if (noteInput) noteInput.value = "";
  updateMarkUnavailableButton();
});
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
