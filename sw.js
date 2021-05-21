const staticAssets=[
  './',
  './static',
  './static/css/materialize.min.css',
  './static/img/icon-192x192.png',
  './static/img/icon-256x256.png',
  './static/img/icon-384x384.png',
  './static/img/icon-512x512.png',
  './static/img/transparent.png',
  './static/js/jquery-3.5.1.min.js',
  './static/js/materialize.min.js',
  './offline.html'
];

self.addEventListener('install', async event=>{
  const cache = await caches.open('static-cache');
  cache.addAll(staticAssets);
});

self.addEventListener('fetch', event => {
  const req = event.request;
  const url = new URL(req.url);

  if(url.origin === location.url){
      event.respondWith(cacheFirst(req));
  } else {
      event.respondWith(networkFirst(req));
  }
});

async function cacheFirst(req){
  const cachedResponse = caches.match(req);
  return cachedResponse || fetch(req);
}

async function networkFirst(req){
  const cache = await caches.open('dynamic-cache');

  try {
      const res = await fetch(req);
      cache.put(req, res.clone());
      return res;
  } catch (error) {
      return await cache.match(req);
  }
}