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
setTimeout(() => {
  apiServer.sendTextContents('/run/434420a4-7440-4da2-acc2-00391b6c5e64/1', 'foo');
}, 10000);
apiServer.on('instantiate', pageId => {
  console.log('instantiate:', pageId);
});
apiServer.on('teardown', pageId => {
  console.log('teardown:', pageId);
});
apiServer.on('buttonClick', widgetId => {
  console.log('button click:', widgetId);
});
apiServer.on('editorContents', (widgetId, contents) => {
  console.log('editor contents:', widgetId, contents);
});
apiServer.on('terminalInput', (widgetId, input) => {
  console.log('terminal input:', widgetId, input);
  apiServer.sendTerminalOutput(widgetId, input);
});
apiServer.listen(8080);
