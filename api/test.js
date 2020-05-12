const ApiServer = require('./ApiServer');
const apiServer = new ApiServer({
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
// apiServer.setPages({
//   '/': {
//     widgets: [
//       { type: 'markdown', content: '# Yo this works!' },
//       { type: 'text', file: '/test.txt' },
//       { type: 'image', file: '/test.png' },
//       { type: 'button', label: 'A button', command: 'date' },
//       { type: 'editor', file: '/test.txt' },
//       { type: 'terminal', workingDirectory: '/test' },
//     ],
//   },
//   '/run': {
//     widgets: [
//       { type: 'markdown', content: '# Yo this works!' },
//       { type: 'text', file: '/test.txt' },
//       { type: 'image', file: '/test.png' },
//       { type: 'button', label: 'A button', command: 'date' },
//       { type: 'editor', file: '/test.txt' },
//       { type: 'terminal', workingDirectory: '/test' },
//     ],
//   },
// });
// setTimeout(() => {
//   apiServer.setPages(apiServer.pages);
// }, 10000);
apiServer.on('instantiate', pageId => {
  console.log('instantiate:', pageId);
});
apiServer.on('teardown', pageId => {
  console.log('teardown:', pageId);
});
apiServer.listen(8080);
