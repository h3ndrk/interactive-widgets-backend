const ApiServer = require('./ApiServer');
const apiServer = new ApiServer();
apiServer.setPages({
  '/': {},
  '/run': {},
});
setTimeout(() => {
  apiServer.setPages({
    '/': {},
    '/run': {},
  });
}, 10000);
apiServer.on('instantiate', (pageUuid, url) => {
  console.log('instantiate:', pageUuid, url);
});
apiServer.on('teardown', (pageUuid, url) => {
  console.log('teardown:', pageUuid, url);
});
apiServer.listen(8080);
