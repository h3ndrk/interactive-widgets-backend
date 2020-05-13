const ids = require('api/ids');

class Volume {
  constructor(docker, pageId) {
    this.docker = docker;
    this.pageId = pageId;
    this.volumeName = `containerized-playground-${ids.idToEncodedId(pageId)}`;
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

module.exports = Volume;
