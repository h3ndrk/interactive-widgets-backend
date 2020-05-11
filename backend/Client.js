class Client {
  constructor(socket) {
    this.socket = socket;
    this.id = socket.id;
    this.pageUuid = null;
    this.url = null;
  }
  sendPages(pages) {
    this.socket.emit('pages', Object.keys(pages));
  }
  sendWidgets(pages) {
    this.socket.emit('widgets', pages[this.url].widgets);
  }
}

module.exports = Client;
