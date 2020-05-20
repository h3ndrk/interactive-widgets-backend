const { spawn } = require('child_process');
const readline = require('readline');
const ids = require('./ids');

class DockerOperator {
  constructor(pages, monitorWriteDirectory) {
    this.pages = pages;
    this.monitorWriteDirectory = monitorWriteDirectory;
    this.sendTextContents = () => { };
    this.sendImageData = () => { };
    this.sendButtonOutput = () => { };
    this.sendEditorContents = () => { };
    this.sendTerminalOutput = () => { };
    this.textContents = {};
    this.buttonOutput = {};
    this.monitorWriteContainers = {};
  }
  async cleanup() {
    // TODO: stop old containers, remove old containers/volumes
  }
  async buildImages() {
    console.log(`Building monitor-write ${this.monitorWriteDirectory} as containerized-playground-monitor-write ...`);
    await new Promise((resolve, reject) => {
      const dockerProcess = spawn('docker', ['build', '--pull', '--tag', 'containerized-playground-monitor-write', this.monitorWriteDirectory]);
      dockerProcess.stdout.pipe(process.stdout);
      dockerProcess.stderr.pipe(process.stderr);
      dockerProcess.on('exit', code => {
        if (code !== 0)
          reject(code);
        else
          resolve();
      });
    });
    for (const url of Object.keys(this.pages)) {
      if (this.pages[url].isInteractive) {
        console.log(`Building page ${this.pages[url].dockerfilePath} as ${this.pages[url].imageName} ...`);
        await new Promise((resolve, reject) => {
          const dockerProcess = spawn('docker', ['build', '--pull', '--tag', this.pages[url].imageName, this.pages[url].basePath]);
          dockerProcess.stdout.pipe(process.stdout);
          dockerProcess.stderr.pipe(process.stderr);
          dockerProcess.on('exit', code => {
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
    const { url, uuid } = ids.pageIdToUrlAndUuid(pageId);
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
      dockerProcess.on('exit', code => {
        if (code !== 0)
          reject(code);
        else
          resolve();
      });
    });
    this.pages[url].widgets.forEach((widget, widgetIndex) => {
      if (widget.isInteractive) {
        switch (widget.type) {
          case "text": {
            const widgetId = ids.urlAndUuidAndWidgetIndexToWidgetId(url, uuid, widgetIndex);
            this.textContents[widgetId] = { contents: '', errors: [] };
            const containerName = `containerized-playground-${ids.idToEncodedId(widgetId)}`;
            console.log(`Running container ${containerName} for widget ${widgetId} (text contents) ...`);
            this.monitorWriteContainers[containerName] = spawn('docker', ['run', '--rm', '--name', containerName, '--network=none', '--mount', `src=${volumeName},dst=/data`, 'containerized-playground-monitor-write', widget.file]);
            const lineStream = readline.createInterface({ input: this.monitorWriteContainers[containerName].stdout, terminal: false });
            lineStream.on('line', async contents => {
              this.textContents[widgetId].contents = contents;
              await this.sendTextContents(widgetId, this.textContents[widgetId]);
            });
            this.monitorWriteContainers[containerName].stderr.on('data', async data => {
              this.textContents[widgetId].errors = [...this.textContents[widgetId].errors.slice(-4), data.toString('base64')];
              await this.sendTextContents(widgetId, this.textContents[widgetId]);
            });
            this.monitorWriteContainers[containerName].on('exit', code => {
              if (code !== 0)
                console.log(`Cannot execute text contents retrieval of widget ${widgetId} (process exited with ${code})`);
              const { [widgetId]: textContentsToRemove, ...textContents } = this.textContents;
              this.textContents = textContents;
            });
            break;
          }
        }
      }
    });
  }
  async stopPage(pageId) {
    const { url, uuid } = ids.pageIdToUrlAndUuid(pageId);
    if (!this.pages[url]) {
      console.log(`Cannot stop page ${pageId} (page not existing)`);
      return;
    }
    const volumeName = `containerized-playground-${ids.idToEncodedId(pageId)}`;
    await Promise.all(this.pages[url].widgets.map(async (widget, widgetIndex) => {
      if (widget.isInteractive) {
        switch (widget.type) {
          case "text": {
            const widgetId = ids.urlAndUuidAndWidgetIndexToWidgetId(url, uuid, widgetIndex);
            const containerName = `containerized-playground-${ids.idToEncodedId(widgetId)}`;
            console.log(`Stopping container ${containerName} for widget ${widgetId} (text contents) ...`);
            const processStopped = new Promise((resolve, _) => {
              this.monitorWriteContainers[containerName].on('exit', () => {
                resolve();
              });
            });
            this.monitorWriteContainers[containerName].stdin.end();
            this.monitorWriteContainers[containerName].kill('SIGTERM');
            await processStopped;
            break;
          }
        }
      }
    }));
    console.log(`Removing volume ${volumeName} for page ${pageId} ...`);
    await new Promise((resolve, reject) => {
      const dockerProcess = spawn('docker', ['volume', 'rm', volumeName]);
      dockerProcess.stdout.pipe(process.stdout);
      dockerProcess.stderr.pipe(process.stderr);
      dockerProcess.on('exit', code => {
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
    const dockerProcess = spawn('docker', ['run', '--rm', '--name', containerName, '--network=none', '--mount', `src=${volumeName},dst=/data`, imageName, 'bash', '-c', this.pages[url].widgets[widgetIndex].command]);
    dockerProcess.stdout.on('data', async data => {
      this.buttonOutput[widgetId] = [...this.buttonOutput[widgetId], { stdout: data.toString('base64') }];
      await this.sendButtonOutput(widgetId, this.buttonOutput[widgetId]);
    });
    dockerProcess.stderr.on('data', async data => {
      this.buttonOutput[widgetId] = [...this.buttonOutput[widgetId], { stderr: data.toString('base64') }];
      await this.sendButtonOutput(widgetId, this.buttonOutput[widgetId]);
    });
    dockerProcess.on('exit', code => {
      if (code !== 0)
        console.log(`Cannot execute button click of widget ${widgetId} (process exited with ${code})`);
      const { [widgetId]: buttonOutputToRemove, ...buttonOutput } = this.buttonOutput;
      this.buttonOutput = buttonOutput;
    });
  }
}

module.exports = DockerOperator;
