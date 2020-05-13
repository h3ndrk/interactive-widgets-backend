const StaticPage = require('./StaticPage');
const ids = require('api/ids');

class InteractivePage extends StaticPage {
  constructor(basePath, url, dockerfilePath) {
    super(basePath, url);
    this.dockerfilePath = dockerfilePath;
    this.imageName = `containerized-playground-${ids.idToEncodedId(url)}`;
  }
  async buildImage(docker) {
    await super.buildImage(docker);
    console.log(`Building page ${this.dockerfilePath} as ${this.imageName} ...`);
    const stream = await this.docker.buildImage({ context: this.basePath }, { t: this.imageName, pull: true });
    this.docker.modem.demuxStream(stream, process.stdout, process.stderr);
    await new Promise((resolve, reject) => {
      this.docker.modem.followProgress(stream, (err, _) => err ? reject(err) : resolve());
    });
  }
}

module.exports = InteractivePage;
