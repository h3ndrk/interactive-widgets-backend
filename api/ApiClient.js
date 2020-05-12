const EventEmitter = require('events');
const EventSource = require('eventsource');

class ApiClient extends EventEmitter {
  constructor(endpoint) {
    this.endpoint = endpoint;
    this.eventSource = new EventSource(endpoint);
    eventSource.addEventListener('server-time', function (event) {
      console.log(event);
      eventSource.close();
    });
  }
}

module.exports = ApiClient;
