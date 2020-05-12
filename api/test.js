const ApiServer = require('./ApiServer');
const apiServer = new ApiServer();
apiServer.setPages({
  '/': {
    widgets: [
      { type: 'markdown', content: '# Yo this works!' },
      { type: 'text', file: '/test.txt' },
      { type: 'image', file: '/test.png' },
      { type: 'button', label: 'A button', command: 'date' },
      { type: 'editor', file: '/test.txt' },
      { type: 'terminal', workingDirectory: '/test' },
    ],
  },
  '/run': {
    widgets: [
      { type: 'markdown', content: '# Yo this works!' },
      { type: 'text', file: '/test.txt' },
      { type: 'image', file: '/test.png' },
      { type: 'button', label: 'A button', command: 'date' },
      { type: 'editor', file: '/test.txt' },
      { type: 'terminal', workingDirectory: '/test' },
    ],
  },
});
setTimeout(() => {
  apiServer.setPages(apiServer.pages);
}, 10000);
apiServer.on('instantiate', (pageUuid, url) => {
  console.log('instantiate:', pageUuid, url);
});
apiServer.on('teardown', (pageUuid, url) => {
  console.log('teardown:', pageUuid, url);
});
apiServer.listen(8080);
