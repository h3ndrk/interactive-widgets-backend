const fs = require('fs');
const path = require('path');
const parse5 = require('parse5');
const ids = require('./ids');

const markdownToWidgets = markdown => {
  const rawBlocks = markdown.split(/((?:\r\n|\r|\n){2,})/);

  const reducedBlocks = (blocks => {
    let resultBlocks = [];
    let currentBlock = [];

    for (const block of blocks) {
      const beginsCode = block.trimLeft().substr(0, 3) === '```';
      const endsCode = block.trimRight().slice(-3) === '```';

      if (beginsCode && !endsCode && currentBlock.length === 0) {
        currentBlock = [block];
      } else if (!beginsCode && endsCode && currentBlock.length > 0) {
        resultBlocks = [...resultBlocks, [...currentBlock, block].join('').trim()];
        currentBlock = [];
      } else if (beginsCode && endsCode && currentBlock.length === 0) {
        resultBlocks = [...resultBlocks, block.trim()];
      } else if (!beginsCode && !endsCode && currentBlock.length > 0) {
        currentBlock = [...currentBlock, block];
      } else if (!beginsCode && !endsCode && currentBlock.length === 0 && block.trim().length > 0) {
        resultBlocks = [...resultBlocks, block.trim()];
      }
    }

    return resultBlocks;
  })(rawBlocks);

  return reducedBlocks.map(block => {
    if ((block.includes('x-text') || block.includes('x-image') || block.includes('x-button') || block.includes('x-editor') || block.includes('x-terminal'))) {
      const widgetElement = parse5.parseFragment(block).childNodes[0];

      if (['x-text', 'x-image', 'x-button', 'x-editor', 'x-terminal'].includes(widgetElement.tagName)) {
        const getAttribute = (element, attribute) => {
          const attr = element.attrs.find(attr => attr.name === attribute);
          if (attr !== undefined && attr.value)
            return attr.value;
          return undefined;
        };

        switch (widgetElement.tagName) {
          case 'x-text':
            return {
              isInteractive: true,
              type: 'text',
              file: getAttribute(widgetElement, 'file'),
            };
          case 'x-image':
            return {
              isInteractive: true,
              type: 'image',
              file: getAttribute(widgetElement, 'file'),
            };
          case 'x-button':
            return {
              isInteractive: true,
              type: 'button',
              label: widgetElement.childNodes[0].value,
              command: getAttribute(widgetElement, 'command'),
            };
          case 'x-editor':
            return {
              isInteractive: true,
              type: 'editor',
              file: getAttribute(widgetElement, 'file'),
            };
          case 'x-terminal':
            return {
              isInteractive: true,
              type: 'terminal',
              workingDirectory: getAttribute(widgetElement, 'working-directory'),
            };
        }
      }
    }

    return {
      isInteractive: false,
      type: 'markdown',
      content: block
    };
  });
};

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