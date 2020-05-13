const { spawn } = require('child_process');
const ids = require('./ids');

class DockerOperator {
  constructor(pages) {
    this.pages = pages;
    this.sendTextContents = () => { };
    this.sendImageData = () => { };
    this.sendButtonOutput = () => { };
    this.sendEditorContents = () => { };
    this.sendTerminalOutput = () => { };
    this.buttonOutput = {};
    // this.volumes = {};
    // this.apiServer.on('instantiate', async pageId => {
    //   console.log('instantiate:', pageId);
    //   this.volumes[pageId] = new Volume(this.docker, pageId);
    //   await this.volumes[pageId].create();
    // });
    // this.apiServer.on('teardown', async pageId => {
    //   console.log('teardown:', pageId);
    //   await this.volumes[pageId].remove();
    //   const { [pageId]: volumeToRemove, ...volumes } = this.volumes;
    //   this.volumes = volumes;
    // });
    // this.apiServer.on('buttonClick', widgetId => {
    //   console.log('button click:', widgetId);
    // });
    // this.apiServer.on('editorContents', (widgetId, contents) => {
    //   console.log('editor contents:', widgetId, contents);
    // });
    // this.apiServer.on('terminalInput', (widgetId, input) => {
    //   console.log('terminal input:', widgetId, input);
    //   this.apiServer.sendTerminalOutput(widgetId, input);
    // });
    // this.clients = {}; // page/UUID hash -> [clients]
    // this.volumes = {}; // page/UUID hash -> volume
    // this.containers = {}; // page/UUID hash -> [container]
  }
  async cleanup() {
    // TODO: stop old containers, remove old containers/volumes
  }
  async buildImages() {
    for (const url of Object.keys(this.pages)) {
      if (this.pages[url].isInteractive) {
        console.log(`Building page ${this.pages[url].dockerfilePath} as ${this.pages[url].imageName} ...`);
        await new Promise((resolve, reject) => {
          const dockerProcess = spawn('docker', ['build', '--pull', '--tag', this.pages[url].imageName, this.pages[url].basePath]);
          dockerProcess.stdout.pipe(process.stdout);
          dockerProcess.stderr.pipe(process.stderr);
          dockerProcess.on('close', code => {
            if (code !== 0)
              reject(code);
            else
              resolve();
          });
        });
      }
    }
  }
  async startPage(pageId) {
    const { url, _ } = ids.pageIdToUrlAndUuid(pageId);
    if (!this.pages[url]) {
      console.log(`Cannot start page ${pageId} (page not existing)`);
      return;
    }
    const volumeName = `containerized-playground-${ids.idToEncodedId(pageId)}`;
    console.log(`Creating volume ${volumeName} for page ${pageId} ...`);
    await new Promise((resolve, reject) => {
      const dockerProcess = spawn('docker', ['volume', 'create', volumeName]);
      dockerProcess.stdout.pipe(process.stdout);
      dockerProcess.stderr.pipe(process.stderr);
      dockerProcess.on('close', code => {
        if (code !== 0)
          reject(code);
        else
          resolve();
      });
    });
  }
  async stopPage(pageId) {
    const { url, _ } = ids.pageIdToUrlAndUuid(pageId);
    if (!this.pages[url]) {
      console.log(`Cannot stop page ${pageId} (page not existing)`);
      return;
    }
    const volumeName = `containerized-playground-${ids.idToEncodedId(pageId)}`;
    console.log(`Removing volume ${volumeName} for page ${pageId} ...`);
    await new Promise((resolve, reject) => {
      const dockerProcess = spawn('docker', ['volume', 'rm', volumeName]);
      dockerProcess.stdout.pipe(process.stdout);
      dockerProcess.stderr.pipe(process.stderr);
      dockerProcess.on('close', code => {
        if (code !== 0)
          reject(code);
        else
          resolve();
      });
    });
  }
  async buttonClick(widgetId) {
    if (this.buttonOutput[widgetId]) {
      console.log(`Cannot execute button click of widget ${widgetId} (click in progress)`);
      return;
    }
    const { url, uuid, widgetIndex } = ids.widgetIdToUrlAndUuidAndWidgetIndex(widgetId);
    if (!this.pages[url]) {
      console.log(`Cannot execute button click on page ${ids.urlAndUuidToPageId(url, uuid)} (page not existing)`);
      return;
    }
    if (widgetIndex < 0 || widgetIndex >= this.pages[url].widgets.length || this.pages[url].widgets[widgetIndex].type !== 'button') {
      console.log(`Cannot execute button click of widget ${widgetId} (widget not existing)`);
      return;
    }
    this.buttonOutput[widgetId] = [];
    const volumeName = `containerized-playground-${ids.idToEncodedId(widgetId)}`;
    const imageName = `containerized-playground-${ids.idToEncodedId(url)}`;
    const containerName = `containerized-playground-${ids.idToEncodedId(widgetId)}`;
    console.log(`Running container ${containerName} for widget ${widgetId} (button click) ...`);
    const dockerProcess = spawn('docker', ['run', '--rm', '--name', containerName, '--network=none', '--mount', `src=${volumeName},dst=/data`, imageName, 'sh', '-c', this.pages[url].widgets[widgetIndex].command]);
    dockerProcess.stdout.on('data', async data => {
      this.buttonOutput[widgetId] = [...this.buttonOutput[widgetId], { stdout: data.toString('base64') }];
      await this.sendButtonOutput(widgetId, this.buttonOutput);
    });
    dockerProcess.stderr.on('data', async data => {
      this.buttonOutput[widgetId] = [...this.buttonOutput[widgetId], { stderr: data.toString('base64') }];
      await this.sendButtonOutput(widgetId, this.buttonOutput);
    });
    dockerProcess.on('close', code => {
      if (code !== 0)
        console.log(`Cannot execute button click of widget ${widgetId} (process exited with ${code})`);
      const { [widgetId]: buttonOutputToRemove, ...buttonOutput } = this.buttonOutput;
      this.buttonOutput = buttonOutput;
    });
  }
  // addClient(socket) {
  //   const client = new Client(socket);
  //   console.log(`+ client ${client.id}`);

  //   client.socket.on('disconnect', async () => {
  //     console.log(`- client ${client.id}`);

  //     if (!client.pageUuid || !client.url)
  //       return;

  //     const pageUuidHash = hashFromPageUuidAndUrl(client.pageUuid, client.url);
  //     if (this.clients[pageUuidHash]) {
  //       this.clients[pageUuidHash] = this.clients[pageUuidHash].filter(c => c.id != client.id);

  //       if (this.clients[pageUuidHash].length === 0) {
  //         const { [pageUuidHash]: clientToRemove, ...clients } = this.clients;
  //         this.clients = clients;
  //         await this.volumes[pageUuidHash].remove();
  //         const { [pageUuidHash]: volumeToRemove, ...volumes } = this.volumes;
  //         this.volumes = volumes;
  //       }
  //     }
  //   });

  //   client.socket.on('request', async request => {
  //     if (!this.pages[request.url]) {
  //       console.warn(`! client ${client.id} (${request.pageUuid}) @ ${request.url} (URL does not exist)`);
  //       return;
  //     }

  //     client.pageUuid = request.pageUuid;
  //     client.url = request.url;
  //     console.log(`  client ${client.id} (${client.pageUuid}) @ ${client.url}`);

  //     const pageUuidHash = hashFromPageUuidAndUrl(client.pageUuid, client.url);
  //     if (!this.clients[pageUuidHash]) {
  //       this.clients[pageUuidHash] = [];
  //       this.volumes[pageUuidHash] = new Volume(this.docker, client.pageUuid, client.url);
  //       await this.volumes[pageUuidHash].create();
  //     }
  //     this.clients[pageUuidHash] = [...this.clients[pageUuidHash], client];

  //     client.sendWidgets(this.pages);
  //   });
  //   client.sendPages(this.pages);
  // }
}

module.exports = DockerOperator;
