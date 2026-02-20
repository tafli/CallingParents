const CACHE_NAME = "calling-parents-v5";
const ASSETS = ["/", "/index.html", "/css/style.css", "/js/i18n.js", "/js/app.js", "/manifest.json", "/lang/de.json", "/lang/en.json"];

// Install: cache app shell
self.addEventListener("install", (event) => {
    event.waitUntil(
        caches.open(CACHE_NAME).then((cache) => cache.addAll(ASSETS))
    );
    self.skipWaiting();
});

// Activate: remove old caches
self.addEventListener("activate", (event) => {
    event.waitUntil(
        caches.keys().then((keys) =>
            Promise.all(
                keys
                    .filter((key) => key !== CACHE_NAME)
                    .map((key) => caches.delete(key))
            )
        )
    );
    self.clients.claim();
});

// Fetch: network-first for API calls, cache-first for app shell
self.addEventListener("fetch", (event) => {
    const url = new URL(event.request.url);

    // API calls: always go to network (they need the server to be reachable)
    if (url.pathname.startsWith("/message/") || url.pathname.startsWith("/children")) {
        event.respondWith(fetch(event.request));
        return;
    }

    // App shell: cache-first, fallback to network
    event.respondWith(
        caches.match(event.request).then((cached) => {
            if (cached) return cached;
            return fetch(event.request).then((response) => {
                // Cache new resources
                if (response.ok) {
                    const clone = response.clone();
                    caches.open(CACHE_NAME).then((cache) => {
                        cache.put(event.request, clone);
                    });
                }
                return response;
            });
        })
    );
});
