// === Storage Keys ===
const STORAGE_CHILDREN = "calling_parents_children";
const STORAGE_SETTINGS = "calling_parents_settings";
const STORAGE_TOKEN = "calling_parents_token";

// === State ===
let children = [];
let settings = { messageId: "Eltern rufen" };
let activeMessage = false;
let authToken = "";

// === Auth Token ===
// Extract token from URL hash fragment (#token=...) and persist in localStorage.
function initToken() {
    const hash = window.location.hash;
    if (hash.startsWith("#token=")) {
        authToken = hash.substring(7);
        localStorage.setItem(STORAGE_TOKEN, authToken);
        // Remove token from URL bar so it's not visible/shared accidentally.
        history.replaceState(null, "", window.location.pathname);
    } else {
        authToken = localStorage.getItem(STORAGE_TOKEN) || "";
    }
}

// Build headers object with auth token included.
function authHeaders(extra = {}) {
    const headers = { ...extra };
    if (authToken) {
        headers["Authorization"] = "Bearer " + authToken;
    }
    return headers;
}

// === DOM Elements ===
const viewMain = document.getElementById("view-main");
const viewSettings = document.getElementById("view-settings");
const btnSettings = document.getElementById("btn-settings");
const btnBack = document.getElementById("btn-back");
const childrenGrid = document.getElementById("children-grid");
const inputName = document.getElementById("input-name");
const btnSend = document.getElementById("btn-send");
const btnClear = document.getElementById("btn-clear");
const statusBar = document.getElementById("status-bar");
const settingMessageId = document.getElementById("setting-message-id");
const btnTestConnection = document.getElementById("btn-test-connection");
const connectionStatus = document.getElementById("connection-status");
const inputAddChild = document.getElementById("input-add-child");
const btnAddChild = document.getElementById("btn-add-child");
const childrenList = document.getElementById("children-list");
const btnReloadChildren = document.getElementById("btn-reload-children");
const toast = document.getElementById("toast");
const headerTitle = document.getElementById("header-title");

// === Initialization ===
function init() {
    initToken();
    loadData();
    renderChildrenGrid();
    renderChildrenList();
    settingMessageId.value = settings.messageId;

    // Fetch server-side children list, then merge
    fetchServerChildren();

    // Event listeners
    btnSettings.addEventListener("click", showSettings);
    btnBack.addEventListener("click", showMain);
    btnSend.addEventListener("click", sendMessage);
    btnClear.addEventListener("click", clearMessage);
    btnTestConnection.addEventListener("click", testConnection);
    btnAddChild.addEventListener("click", addChild);
    btnReloadChildren.addEventListener("click", reloadChildren);
    inputName.addEventListener("input", onNameInput);
    inputName.addEventListener("keydown", (e) => {
        if (e.key === "Enter" && inputName.value.trim()) sendMessage();
    });
    inputAddChild.addEventListener("keydown", (e) => {
        if (e.key === "Enter") addChild();
    });
    settingMessageId.addEventListener("change", () => {
        settings.messageId = settingMessageId.value.trim() || "Eltern rufen";
        saveSettings();
    });
}

// === Data Persistence ===
function loadData() {
    try {
        const storedChildren = localStorage.getItem(STORAGE_CHILDREN);
        if (storedChildren) children = JSON.parse(storedChildren);
    } catch (_) {
        children = [];
    }
    try {
        const storedSettings = localStorage.getItem(STORAGE_SETTINGS);
        if (storedSettings) settings = { ...settings, ...JSON.parse(storedSettings) };
    } catch (_) {
        // keep defaults
    }
}

function saveChildren() {
    children.sort((a, b) => a.localeCompare(b, "de"));
    localStorage.setItem(STORAGE_CHILDREN, JSON.stringify(children));
}

function saveSettings() {
    localStorage.setItem(STORAGE_SETTINGS, JSON.stringify(settings));
}

// === Server Children Sync ===
async function fetchServerChildren() {
    try {
        const resp = await fetch("/children", {
            headers: authHeaders(),
        });
        if (!resp.ok) return;

        const serverNames = await resp.json();
        if (!Array.isArray(serverNames) || serverNames.length === 0) return;

        // Merge: add server names not already in the local list
        let changed = false;
        for (const name of serverNames) {
            if (!children.includes(name)) {
                children.push(name);
                changed = true;
            }
        }

        if (changed) {
            saveChildren();
            renderChildrenGrid();
            renderChildrenList();
        }
    } catch (_) {
        // Offline or server unreachable — keep local list
    }
}

// Full replace of local list with server list.
async function reloadChildren() {
    try {
        const resp = await fetch("/children", {
            headers: authHeaders(),
        });
        if (!resp.ok) {
            showToast("Serverliste konnte nicht geladen werden", "error");
            return;
        }

        const serverNames = await resp.json();
        if (!Array.isArray(serverNames)) {
            showToast("Ungültige Antwort vom Server", "error");
            return;
        }

        children = serverNames;
        saveChildren();
        renderChildrenGrid();
        renderChildrenList();
        showToast(`${children.length} Namen vom Server geladen`, "success");
    } catch (_) {
        showToast("Server nicht erreichbar", "error");
    }
}

// === View Switching ===
function showSettings() {
    viewMain.classList.add("hidden");
    viewSettings.classList.remove("hidden");
    headerTitle.textContent = "Einstellungen";
    btnSettings.classList.add("hidden");
}

function showMain() {
    viewSettings.classList.add("hidden");
    viewMain.classList.remove("hidden");
    headerTitle.textContent = "Eltern rufen";
    btnSettings.classList.remove("hidden");
    renderChildrenGrid();
}

// === Children Grid (Main View) ===
function renderChildrenGrid() {
    childrenGrid.innerHTML = "";
    children.forEach((name) => {
        const btn = document.createElement("button");
        btn.className = "child-btn";
        btn.textContent = name;
        btn.addEventListener("click", () => selectChild(name));
        childrenGrid.appendChild(btn);
    });
}

function selectChild(name) {
    inputName.value = name;
    onNameInput();

    // Highlight the selected button
    document.querySelectorAll(".child-btn").forEach((btn) => {
        btn.classList.toggle("selected", btn.textContent === name);
    });
}

function onNameInput() {
    btnSend.disabled = !inputName.value.trim();

    // Update button highlights based on current input
    const currentName = inputName.value.trim();
    document.querySelectorAll(".child-btn").forEach((btn) => {
        btn.classList.toggle("selected", btn.textContent === currentName);
    });
}

// === Children List (Settings View) ===
function renderChildrenList() {
    childrenList.innerHTML = "";
    children.forEach((name, index) => {
        const li = document.createElement("li");

        const span = document.createElement("span");
        span.textContent = name;

        const removeBtn = document.createElement("button");
        removeBtn.className = "btn-remove";
        removeBtn.textContent = "✕";
        removeBtn.setAttribute("aria-label", `${name} entfernen`);
        removeBtn.addEventListener("click", () => removeChild(index));

        li.appendChild(span);
        li.appendChild(removeBtn);
        childrenList.appendChild(li);
    });
}

function addChild() {
    const name = inputAddChild.value.trim();
    if (!name) return;
    if (children.includes(name)) {
        showToast(`"${name}" ist bereits vorhanden`, "error");
        return;
    }

    children.push(name);
    saveChildren();
    renderChildrenList();
    inputAddChild.value = "";
    inputAddChild.focus();

    // Persist to server-side children file (fire-and-forget).
    fetch("/children", {
        method: "POST",
        headers: authHeaders({ "Content-Type": "application/json" }),
        body: JSON.stringify({ name }),
    }).catch(() => {
        // Server sync is best-effort; localStorage is the primary store.
    });
}

function removeChild(index) {
    const name = children[index];
    children.splice(index, 1);
    saveChildren();
    renderChildrenList();

    // Sync deletion to server (fire-and-forget).
    if (name) {
        fetch("/children", {
            method: "DELETE",
            headers: authHeaders({ "Content-Type": "application/json" }),
            body: JSON.stringify({ name }),
        }).catch(() => {
            // Server sync is best-effort.
        });
    }
}

// === ProPresenter API ===
async function sendMessage() {
    const name = inputName.value.trim();
    if (!name) return;

    btnSend.disabled = true;

    try {
        const messageId = encodeURIComponent(settings.messageId);
        const resp = await fetch(`/api/v1/message/${messageId}/trigger`, {
            method: "POST",
            headers: authHeaders({ "Content-Type": "application/json" }),
            body: JSON.stringify([
                {
                    name: "Name",
                    text: { text: name },
                },
            ]),
        });

        if (!resp.ok && resp.status !== 204) {
            throw new Error(`HTTP ${resp.status}`);
        }

        activeMessage = true;
        showStatus(`Anzeige: "Eltern von ${name}"`, "active");
        showToast("Nachricht gesendet", "success");
    } catch (err) {
        showToast(`Fehler: ${err.message}`, "error");
        showStatus("Senden fehlgeschlagen", "error");
    } finally {
        btnSend.disabled = !inputName.value.trim();
    }
}

async function clearMessage() {
    try {
        const messageId = encodeURIComponent(settings.messageId);
        const resp = await fetch(`/api/v1/message/${messageId}/clear`, {
            headers: authHeaders(),
        });

        if (!resp.ok && resp.status !== 204) {
            throw new Error(`HTTP ${resp.status}`);
        }

        activeMessage = false;
        hideStatus();
        showToast("Nachricht gelöscht", "success");

        // Reset selection
        inputName.value = "";
        onNameInput();
    } catch (err) {
        showToast(`Fehler: ${err.message}`, "error");
    }
}

async function testConnection() {
    connectionStatus.textContent = "Teste Verbindung…";
    connectionStatus.className = "connection-status";

    try {
        const resp = await fetch("/api/v1/messages", {
            headers: authHeaders(),
        });
        if (!resp.ok) throw new Error(`HTTP ${resp.status}`);

        const data = await resp.json();
        const count = Array.isArray(data) ? data.length : 0;
        connectionStatus.textContent = `Verbunden — ${count} Nachricht(en) gefunden`;
        connectionStatus.className = "connection-status success";
    } catch (err) {
        connectionStatus.textContent = `Verbindung fehlgeschlagen: ${err.message}`;
        connectionStatus.className = "connection-status error";
    }
}

// === Status Bar ===
function showStatus(text, type) {
    statusBar.textContent = text;
    statusBar.className = `status-bar ${type}`;
    statusBar.classList.remove("hidden");
}

function hideStatus() {
    statusBar.classList.add("hidden");
    statusBar.className = "status-bar hidden";
}

// === Toast ===
let toastTimeout = null;

function showToast(message, type) {
    if (toastTimeout) clearTimeout(toastTimeout);

    toast.textContent = message;
    toast.className = `toast ${type}`;

    toastTimeout = setTimeout(() => {
        toast.classList.add("hidden");
    }, 3000);
}

// === Service Worker Registration ===
if ("serviceWorker" in navigator) {
    navigator.serviceWorker.register("sw.js").catch((err) => {
        console.warn("Service worker registration failed:", err);
    });
}

// === Start ===
init();
