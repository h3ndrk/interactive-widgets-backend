const hashFromPageUuidAndUrl = require('./hashFromPageUuidAndUrl');
const Client = require('./Client');
const Volume = require('./Volume');

class DockerBackend {
  constructor(docker, pages) {
    this.docker = docker;
    this.pages = pages;
    this.clients = {}; // page/UUID hash -> [clients]
    this.volumes = {}; // page/UUID hash -> volume
    this.containers = {}; // page/UUID hash -> [container]
  }
  async buildPages() {
    for (const url of Object.keys(this.pages)) {
      await this.pages[url].buildImage(this.docker);
    }
  }
  addClient(socket) {
    const client = new Client(socket);
    console.log(`+ client ${client.id}`);

    client.socket.on('disconnect', async () => {
      console.log(`- client ${client.id}`);

      if (!client.pageUuid || !client.url)
        return;

      const pageUuidHash = hashFromPageUuidAndUrl(client.pageUuid, client.url);
      if (this.clients[pageUuidHash]) {
        this.clients[pageUuidHash] = this.clients[pageUuidHash].filter(c => c.id != client.id);

        if (this.clients[pageUuidHash].length === 0) {
          const { [pageUuidHash]: clientToRemove, ...clients } = this.clients;
          this.clients = clients;
          await this.volumes[pageUuidHash].remove();
          const { [pageUuidHash]: volumeToRemove, ...volumes } = this.volumes;
          this.volumes = volumes;
        }
      }
    });

    client.socket.on('request', async request => {
      if (!this.pages[request.url]) {
        console.warn(`! client ${client.id} (${request.pageUuid}) @ ${request.url} (URL does not exist)`);
        return;
      }

      client.pageUuid = request.pageUuid;
      client.url = request.url;
      console.log(`  client ${client.id} (${client.pageUuid}) @ ${client.url}`);

      const pageUuidHash = hashFromPageUuidAndUrl(client.pageUuid, client.url);
      if (!this.clients[pageUuidHash]) {
        this.clients[pageUuidHash] = [];
        this.volumes[pageUuidHash] = new Volume(this.docker, client.pageUuid, client.url);
        await this.volumes[pageUuidHash].create();
      }
      this.clients[pageUuidHash] = [...this.clients[pageUuidHash], client];

      client.sendWidgets(this.pages);
    });
    client.sendPages(this.pages);
  }
}

module.exports = DockerBackend;
