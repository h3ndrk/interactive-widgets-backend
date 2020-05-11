const io = require('socket.io-client');
const socket = io('http://localhost:3001');

socket.on('connect', () => {
  console.log('connected');
  // socket.emit('welcome', 'welcome man');
});
socket.on('disconnect', () => {
  console.log('disconnected');
})
socket.on('pages', pages => {
  console.log('pages:', pages);
  socket.emit('request', {pageUuid: 'page-uuid', url: '/'});
});
