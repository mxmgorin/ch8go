self.addEventListener("fetch", event => {
  event.respondWith(
    caches.open("ch8go-runtime").then(cache =>
      cache.match(event.request).then(resp => {
        const fetchPromise = fetch(event.request).then(networkResp => {
          cache.put(event.request, networkResp.clone());
          return networkResp;
        });
        return resp || fetchPromise;
      })
    )
  );
});
