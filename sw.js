const staticAssets=[
  './',
  './manifest.json',
  './static/css/materialize.min.css',
  './static/img/icon-192x192.png',
  './static/img/icon-256x256.png',
  './static/img/icon-384x384.png',
  './static/img/icon-512x512.png',
  './static/js/jquery-3.5.1.min.js',
  './static/js/materialize.min.js',
];

self.addEventListener('install', async event=>{
  event.waitUntil(
    caches.open('static-cache')
      .then(function(cache) {
        console.log('Opened cache');
        return cache.addAll(staticAssets);
      })
  );
});

self.addEventListener('fetch', event => {  
  event.respondWith(cacheFirst(event.request));
});

async function cacheFirst(req){
  const cachedResponse = caches.match(req);
  return cachedResponse || fetch(req);
}