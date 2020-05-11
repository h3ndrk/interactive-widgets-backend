const io = require('socket.io');
const server = io.listen(3001);

server.on('connection', socket => {
  console.log('user connected');
  socket.emit('welcome', 'welcome man');
});
