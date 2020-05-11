const hashFromPageUuidAndUrl = require('./hashFromPageUuidAndUrl');

class LongRunningContainer {
  constructor(docker, pageUuid, url) {
    this.docker = docker;
    this.pageUuid = pageUuid;
    this.url = url;
    this.hash = hashFromPageUuidAndUrl(this.pageUuid, this.url);
    this.volumeName = `containerized-playground-${this.hash}`;
  }
  async create() {
    console.log(`Creating volume ${this.volumeName} ...`);
    this.volume = await this.docker.createVolume({ name: this.volumeName });
  }
  async remove() {
    console.log(`Removing volume ${this.volumeName} ...`);
    await this.volume.remove();
  }
}

module.exports = LongRunningContainer;
