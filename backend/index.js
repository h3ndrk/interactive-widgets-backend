// const io = require('socket.io');
// const server = io.listen(3001);

// server.on('connection', socket => {
//   console.log('user connected');
//   socket.emit('welcome', 'welcome man');
// });

const { program } = require('commander');
const fs = require('fs');
const path = require('path');
const crypto = require('crypto');
const Docker = require('dockerode');
const io = require('socket.io');
// const markdown2Widgets = require('./markdown2Widgets');

// console.log(markdown2Widgets(markdown));

class StaticPage {
  constructor(basePath, url) {
    this.basePath = basePath;
    this.url = url;
  }
  async buildImage(docker) {
    this.docker = docker;
  }
}

class InteractivePage extends StaticPage {
  constructor(basePath, url, dockerfilePath) {
    super(basePath, url);
    this.dockerfilePath = dockerfilePath;
    this.hash = crypto.createHash('sha256').update(this.url).digest('hex');
    this.imageName = `containerized-playground-${this.hash}`;
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

async function* walk(directory) {
  for await (const currentPath of await fs.promises.opendir(directory)) {
    const entry = path.join(directory, currentPath.name);
    if (currentPath.isDirectory())
      yield* walk(entry);
    else if (currentPath.isFile())
      yield entry;
  }
}

async function* getPages(pagesDirectory) {
  for await (const p of walk(pagesDirectory)) {
    if (path.basename(p) === 'page.md') {
      const pageBasePath = path.dirname(p);
      // generate URL '/some/page' from page path '.../pages/some/page'
      const url = path.join(path.sep, path.relative(pagesDirectory, pageBasePath));
      try {
        const dockerfilePath = path.join(pageBasePath, 'Dockerfile');
        await fs.promises.access(dockerfilePath, fs.F_OK);
        yield new InteractivePage(pageBasePath, url, dockerfilePath);
      } catch (e) {
        yield new StaticPage(pageBasePath, url);
      }
    }
  }
}

class Client {
  constructor(socket) {
    this.socket = socket;
    this.id = socket.id;
    this.pageUuid = null;
    this.url = null;
  }
  sendPages(pages) {
    this.socket.emit('pages', pages.map(page => page.url));
  }
}

class Volume {
  constructor(docker, pageUuid, url) {
    this.docker = docker;
    this.pageUuid = pageUuid;
    this.url = url;
    this.hash = Volume.hashFromPageUuidAndUrl(this.pageUuid, this.url);
    this.volumeName = `containerized-playground-${this.hash}`;
    this.numberOfClients = 0;
  }
  async registerClient() {
    this.numberOfClients++;
    if (this.numberOfClients === 1) {
      await this.create();
    }
  }
  async deregisterClient() {
    this.numberOfClients--;
    if (this.numberOfClients === 0) {
      await this.remove();
    }
  }
  isOrphan() {
    return this.numberOfClients === 0;
  }
  async create() {
    console.log(`Creating volume ${this.volumeName} ...`);
    this.volume = await this.docker.createVolume({ name: this.volumeName });
  }
  async remove() {
    console.log(`Removing volume ${this.volumeName} ...`);
    await this.volume.remove();
  }
  static hashFromPageUuidAndUrl(pageUuid, url) {
    return crypto.createHash('sha256').update(`${pageUuid} ${url}`).digest('hex');
  }
}

class DockerBackend {
  constructor(docker, pages) {
    this.docker = docker;
    this.pages = pages;
    this.clients = [];
    this.volumes = {};
    // this.containers = [];
  }
  async buildPages() {
    for (const page of this.pages) {
      await page.buildImage(this.docker);
    }
  }
  addClient(socket) {
    const client = new Client(socket);
    this.clients = [...this.clients, client];
    console.log(`+ client ${client.id}`);

    client.socket.on('disconnect', async () => {
      console.log(`- client ${client.id}`);
      this.clients = this.clients.filter(c => c.id != client.id);

      const volumeHash = Volume.hashFromPageUuidAndUrl(client.pageUuid, client.url);
      await this.volumes[volumeHash].deregisterClient();
      if (this.volumes[volumeHash].isOrphan())
        delete this.volumes[volumeHash];
    });

    client.socket.on('request', async request => {
      if (!this.pages.find(page => page.url === request.url)) {
        console.warn(`! client ${client.id} (${request.pageUuid}) @ ${request.url} (URL does not exist)`);
        return;
      }
      client.pageUuid = request.pageUuid;
      client.url = request.url;
      console.log(`  client ${client.id} (${client.pageUuid}) @ ${client.url}`);

      const volumeHash = Volume.hashFromPageUuidAndUrl(client.pageUuid, client.url);
      if (!this.volumes[volumeHash])
        this.volumes[volumeHash] = new Volume(this.docker, client.pageUuid, client.url);
      await this.volumes[volumeHash].registerClient();
    });
    client.sendPages(this.pages);
  }
}

async function main() {
  let pagesDirectory;
  program.arguments('<pages>').action(pagesValue => {
    pagesDirectory = pagesValue;
  });
  program.parse(process.argv);
  const docker = new Docker();

  let pages = [];
  for await (const page of getPages(pagesDirectory)) {
    pages = [...pages, page];
  }

  const dockerBackend = new DockerBackend(docker, pages);
  await dockerBackend.buildPages();

  console.log('Listening ...');
  io.listen(3001).on('connect', dockerBackend.addClient.bind(dockerBackend));
}

main()
