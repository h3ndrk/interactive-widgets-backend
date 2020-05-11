// const io = require('socket.io');
// const server = io.listen(3001);

// server.on('connection', socket => {
//   console.log('user connected');
//   socket.emit('welcome', 'welcome man');
// });

const parse5 = require('parse5');

const markdown = `
# Hello \`World!\`

\`\`\`dockerfile
FROM ubuntu:18.04

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install --no-install-recommends -y \\
    iproute2 inotify-tools \\
    && rm -Rf /var/lib/apt/lists/*

CMD tail -f /dev/null
\`\`\`

## C++ code

\`\`\`cpp
class Foo {
public:
    explicit Foo(const std::string &temp);
};
\`\`\`

## Image

![The Office](source.gif)

## Widget Tests


Here come widgets:




<x-text file="/test.txt" />

<x-image file="/image.png" />

<x-button command="echo &quot;The current server time is: $(date)&quot;">Print Hello World</x-button>

<x-editor type="text" file="/test.txt" />

<x-terminal working-directory="/" />
`

const markdown2Widgets = markdown => {
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
              'type': 'text',
              'file': getAttribute(widgetElement, 'file'),
            };
          case 'x-image':
            return {
              'type': 'image',
              'file': getAttribute(widgetElement, 'file'),
            };
          case 'x-button':
            return {
              'type': 'button',
              'label': widgetElement.childNodes[0].value,
              'command': getAttribute(widgetElement, 'command'),
            };
          case 'x-editor':
            return {
              'type': 'editor',
              'file': getAttribute(widgetElement, 'file'),
            };
          case 'x-terminal':
            return {
              'type': 'terminal',
              'workingDirectory': getAttribute(widgetElement, 'working-directory'),
            };
        }
      }
    }
    return { 'type': 'markdown', 'content': block };
  });
};

console.log(markdown2Widgets(markdown));
