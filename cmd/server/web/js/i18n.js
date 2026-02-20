// === Internationalization (i18n) ===
// Lightweight translation module. No dependencies.
// Translations are loaded from /lang/{code}.json files.
// Language detection: localStorage override > navigator.language > "de" fallback.

const STORAGE_LANG = "calling_parents_lang";
const FALLBACK_LANG = "de";

// Available languages for the UI picker.
const availableLanguages = [
    { code: "de", label: "Deutsch" },
    { code: "en", label: "English" },
];

const translations = {};
let currentLang = FALLBACK_LANG;

// Load a language file from the server. Returns the translations object or null.
async function loadLanguage(lang) {
    if (translations[lang]) return translations[lang];
    try {
        const resp = await fetch(`/lang/${lang}.json`);
        if (!resp.ok) return null;
        translations[lang] = await resp.json();
        return translations[lang];
    } catch (_) {
        return null;
    }
}

function detectLanguage() {
    const stored = localStorage.getItem(STORAGE_LANG);
    if (stored && availableLanguages.some((l) => l.code === stored)) return stored;

    const nav = (navigator.language || FALLBACK_LANG).split("-")[0].toLowerCase();
    return availableLanguages.some((l) => l.code === nav) ? nav : FALLBACK_LANG;
}

async function setLanguage(lang) {
    await loadLanguage(lang);
    currentLang = lang;
    localStorage.setItem(STORAGE_LANG, lang);
    document.documentElement.lang = lang;
}

async function initI18n() {
    const detected = detectLanguage();
    // Always load the fallback language first.
    await loadLanguage(FALLBACK_LANG);
    if (detected !== FALLBACK_LANG) {
        await loadLanguage(detected);
    }
    currentLang = detected;
    document.documentElement.lang = currentLang;
}

// Translate a key with optional parameter substitution.
// Usage: t("toast.sent", { name: "Paul" }) → "Nachricht gesendet: Paul ✓"
function t(key, params = {}) {
    let text = (translations[currentLang] && translations[currentLang][key])
        || (translations[FALLBACK_LANG] && translations[FALLBACK_LANG][key])
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
