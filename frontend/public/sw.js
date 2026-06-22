const CACHE_NAME = "keep-v1"

const PRECACHE_URLS = [
  "/",
  "/archive",
  "/trash",
  "/reminders",
  "/labels",
  "/icon-192x192.png",
  "/icon-512x512.png",
]

self.addEventListener("install", (event) => {
  event.waitUntil(
    caches.open(CACHE_NAME).then((cache) => cache.addAll(PRECACHE_URLS))
  )
  self.skipWaiting()
})

self.addEventListener("activate", (event) => {
  event.waitUntil(
    caches.keys().then((keys) =>
      Promise.all(
        keys.filter((k) => k !== CACHE_NAME).map((k) => caches.delete(k))
      )
    )
  )
  self.clients.claim()
})

self.addEventListener("fetch", (event) => {
  const { request } = event
  const url = new URL(request.url)

  // Skip non-GET and API/auth requests
  if (request.method !== "GET") return
  if (url.pathname.startsWith("/v1/")) return
  if (url.pathname.startsWith("/oauth2/")) return
  if (url.pathname.startsWith("/openapi")) return

  // Network-first for HTML navigation, cache-first for assets
  if (request.mode === "navigate") {
    event.respondWith(
      fetch(request)
        .then((response) => {
          const clone = response.clone()
          caches.open(CACHE_NAME).then((cache) => cache.put(request, clone))
          return response
        })
        .catch(() => caches.match(request).then((cached) => cached || caches.match("/")))
    )
    return
  }

  // Cache-first for static assets
  if (
    url.pathname.startsWith("/_next/static/") ||
    url.pathname.match(/\.(png|svg|ico|woff2?|css|js)$/)
  ) {
    event.respondWith(
      caches.match(request).then(
        (cached) =>
          cached ||
          fetch(request).then((response) => {
            const clone = response.clone()
            caches.open(CACHE_NAME).then((cache) => cache.put(request, clone))
            return response
          })
      )
    )
    return
  }
})
