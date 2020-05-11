// const io = require('socket.io');
// const server = io.listen(3001);

// server.on('connection', socket => {
//   console.log('user connected');
//   socket.emit('welcome', 'welcome man');
// });

const parse5 = require('parse5');
const { program } = require('commander');
const fs = require('fs');
const path = require('path');
// const markdown2Widgets = require('./markdown2Widgets');

// console.log(markdown2Widgets(markdown));

let pages;
program.arguments('<pages>').action(pagesValue => {
  pages = pagesValue;
});
program.parse(process.argv);

console.log(pages);

async function* walk(dir) {
  for await (const d of await fs.promises.opendir(dir)) {
    const entry = path.join(dir, d.name);
    if (d.isDirectory()) yield* walk(entry);
    else if (d.isFile()) yield entry;
  }
}

async function getPages(pages) {
  for await (const p of walk(pages)) {
    console.log(p);
  }
}

getPages(pages);
