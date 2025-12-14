self.addEventListener("fetch", (event) => {
  const url = new URL(event.request.url);

  if (url.protocol !== "http:" && url.protocol !== "https:") {
    return;
  }

  event.respondWith(
    caches.match(event.request).then((resp) => {
      return (
        resp ||
        fetch(event.request).then(async (networkResp) => {
          const cache = await caches.open("ch8go-runtime");
          cache.put(event.request, networkResp.clone());
          return networkResp;
        })
      );
    }),
  );
});
