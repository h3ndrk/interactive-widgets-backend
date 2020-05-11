const fs = require('fs');
const path = require('path');
const StaticPage = require('./StaticPage');
const InteractivePage = require('./InteractivePage');

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

module.exports = getPages;
