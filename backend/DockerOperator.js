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
}

module.exports = DockerOperator;
