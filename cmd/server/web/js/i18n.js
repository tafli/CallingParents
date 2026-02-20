// === Internationalization (i18n) ===
// Lightweight translation module. No dependencies.
// Language detection: localStorage override > navigator.language > "de" fallback.

const STORAGE_LANG = "calling_parents_lang";

const translations = {
    de: {
        // App title
        "app.title": "Eltern rufen",
        "app.description": "Eltern über ProPresenter aufrufen",

        // Header
        "header.settings": "Einstellungen",
        "header.statusDot": "Verbindungsstatus",

        // Main view
        "connection.banner": "⚠ ProPresenter nicht erreichbar",
        "input.placeholder": "Name eingeben…",
        "input.clearLabel": "Eingabe löschen",
        "btn.send": "Senden",

        // Settings view
        "settings.title": "Einstellungen",
        "settings.connection": "Verbindung",
        "settings.testConnection": "Verbindung testen",
        "settings.manageChildren": "Kinder verwalten",
        "settings.addPlaceholder": "Name hinzufügen…",
        "settings.reloadFromServer": "Liste vom Server laden",
        "settings.back": "Zurück",
        "settings.language": "Sprache",

        // Connection status
        "connection.testing": "Teste Verbindung…",
        "connection.success": "Verbunden — {count} Nachricht(en) gefunden",
        "connection.failed": "Verbindung fehlgeschlagen: {error}",
        "connection.connected": "ProPresenter verbunden",
        "connection.disconnected": "ProPresenter nicht erreichbar",

        // Toasts / messages
        "toast.sent": "Nachricht gesendet: {name} ✓",
        "toast.sendFailed": "Fehler: {error}",
        "toast.cleared": "Nachricht gelöscht",
        "toast.autoCleared": "Nachricht automatisch gelöscht",
        "toast.autoClearFailed": "Auto-Löschen fehlgeschlagen: {error}",
        "toast.childExists": "\"{name}\" ist bereits vorhanden",
        "toast.serverListLoaded": "{count} Namen vom Server geladen",
        "toast.serverListFailed": "Serverliste konnte nicht geladen werden",
        "toast.serverInvalidResponse": "Ungültige Antwort vom Server",
        "toast.serverUnreachable": "Server nicht erreichbar",

        // Status bar
        "status.showing": "Anzeige: \"Eltern von {name}\"",
        "status.sendFailed": "Senden fehlgeschlagen",

        // Auth error
        "auth.title": "Nicht autorisiert",
        "auth.message": "Bitte scanne den QR-Code erneut, um Zugang zu erhalten.",

        // Aria labels
        "aria.removeChild": "{name} entfernen",

        // Children grid empty state (CSS-driven, not used in JS)
        "grid.empty": "Keine Kinder eingetragen. Öffne die Einstellungen (⚙), um Namen hinzuzufügen.",
    },

    en: {
        "app.title": "Call Parents",
        "app.description": "Call parents via ProPresenter",

        "header.settings": "Settings",
        "header.statusDot": "Connection status",

        "connection.banner": "⚠ ProPresenter not reachable",
        "input.placeholder": "Enter name…",
        "input.clearLabel": "Clear input",
        "btn.send": "Send",

        "settings.title": "Settings",
        "settings.connection": "Connection",
        "settings.testConnection": "Test connection",
        "settings.manageChildren": "Manage children",
        "settings.addPlaceholder": "Add name…",
        "settings.reloadFromServer": "Reload list from server",
        "settings.back": "Back",
        "settings.language": "Language",

        "connection.testing": "Testing connection…",
        "connection.success": "Connected — {count} message(s) found",
        "connection.failed": "Connection failed: {error}",
        "connection.connected": "ProPresenter connected",
        "connection.disconnected": "ProPresenter not reachable",

        "toast.sent": "Message sent: {name} ✓",
        "toast.sendFailed": "Error: {error}",
        "toast.cleared": "Message cleared",
        "toast.autoCleared": "Message auto-cleared",
        "toast.autoClearFailed": "Auto-clear failed: {error}",
        "toast.childExists": "\"{name}\" already exists",
        "toast.serverListLoaded": "{count} names loaded from server",
        "toast.serverListFailed": "Could not load server list",
        "toast.serverInvalidResponse": "Invalid server response",
        "toast.serverUnreachable": "Server not reachable",

        "status.showing": "Showing: \"Parents of {name}\"",
        "status.sendFailed": "Send failed",

        "auth.title": "Not authorized",
        "auth.message": "Please scan the QR code again to get access.",

        "aria.removeChild": "Remove {name}",

        "grid.empty": "No children added. Open settings (⚙) to add names.",
    },
};

// Available languages for the UI picker.
const availableLanguages = [
    { code: "de", label: "Deutsch" },
    { code: "en", label: "English" },
];

let currentLang = "de";

function detectLanguage() {
    const stored = localStorage.getItem(STORAGE_LANG);
    if (stored && translations[stored]) return stored;

    const nav = (navigator.language || "de").split("-")[0].toLowerCase();
    return translations[nav] ? nav : "de";
}

function setLanguage(lang) {
    if (!translations[lang]) return;
    currentLang = lang;
    localStorage.setItem(STORAGE_LANG, lang);
    document.documentElement.lang = lang;
}

function initI18n() {
    currentLang = detectLanguage();
    document.documentElement.lang = currentLang;
}

// Translate a key with optional parameter substitution.
// Usage: t("toast.sent", { name: "Paul" }) → "Nachricht gesendet: Paul ✓"
function t(key, params = {}) {
    let text = (translations[currentLang] && translations[currentLang][key])
        || (translations.de && translations.de[key])
        || key;
    for (const [k, v] of Object.entries(params)) {
        text = text.replace(`{${k}}`, v);
    }
    return text;
}

// Apply translations to static HTML elements with data-i18n attributes.
function applyI18nToDOM() {
    document.querySelectorAll("[data-i18n]").forEach((el) => {
        el.textContent = t(el.dataset.i18n);
    });
    document.querySelectorAll("[data-i18n-placeholder]").forEach((el) => {
        el.placeholder = t(el.dataset.i18nPlaceholder);
    });
    document.querySelectorAll("[data-i18n-aria]").forEach((el) => {
        el.setAttribute("aria-label", t(el.dataset.i18nAria));
    });
    document.querySelectorAll("[data-i18n-title]").forEach((el) => {
        el.title = t(el.dataset.i18nTitle);
    });
    // Page title
    document.title = t("app.title");
}
