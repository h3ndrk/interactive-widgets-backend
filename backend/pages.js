const fs = require('fs');
const path = require('path');
const markdownToWidgets = require('./markdownToWidgets');
const ids = require('./ids');

async function* walk(directory) {
  for await (const currentPath of await fs.promises.opendir(directory)) {
    const entry = path.join(directory, currentPath.name);
    if (currentPath.isDirectory())
      yield* walk(entry);
    else if (currentPath.isFile())
      yield entry;
  }
}

module.exports = async pagesDirectory => {
  let pages = {};
  for await (const p of walk(pagesDirectory)) {
    if (path.basename(p) === 'page.md') {
      const pageBasePath = path.dirname(p);
      // generate URL '/some/page' from page path '.../pages/some/page'
      const url = path.join(path.sep, path.relative(pagesDirectory, pageBasePath));
      const page = {
        isInteractive: false,
        basePath: pageBasePath,
        url: url,
        widgets: markdownToWidgets((await fs.promises.readFile(path.join(pageBasePath, 'page.md'))).toString('utf8')),
      };
      if (page.widgets.some(widget => widget.isInteractive)) {
        const dockerfilePath = path.join(pageBasePath, 'Dockerfile');
        await fs.promises.access(dockerfilePath, fs.F_OK);
        pages = {
          ...pages,
          [url]: {
            ...page,
            isInteractive: true,
            dockerfilePath: dockerfilePath,
            imageName: `containerized-playground-${ids.idToEncodedId(url)}`,
          },
        };
      } else {
        pages = {
          ...pages,
          [url]: page,
        };
      }
    }
  }
  return pages;
};
