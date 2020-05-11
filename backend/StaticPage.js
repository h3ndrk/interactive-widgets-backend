const fs = require('fs');
const path = require('path');
const markdown2Widgets = require('./markdown2Widgets');

class StaticPage {
  constructor(basePath, url) {
    this.basePath = basePath;
    this.url = url;
    this.widgets = [];
  }
  async readPage() {
    this.widgets = markdown2Widgets((await fs.promises.readFile(path.join(this.basePath, 'page.md'))).toString('utf8'));
  }
  async buildImage(docker) {
    this.docker = docker;
  }
}

module.exports = StaticPage;
