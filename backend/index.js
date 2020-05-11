const { program } = require('commander');
const Docker = require('dockerode');
const io = require('socket.io');
const getPages = require('./getPages');
const DockerBackend = require('./DockerBackend');

async function main() {
  let pagesDirectory;
  program.arguments('<pages>').action(pagesValue => {
    pagesDirectory = pagesValue;
  });
  program.parse(process.argv);
  const docker = new Docker();

  let pages = {};
  for await (const page of getPages(pagesDirectory)) {
    await page.readPage();
    pages = { ...pages, [page.url]: page };
  }

  const dockerBackend = new DockerBackend(docker, pages);
  await dockerBackend.buildPages();

  console.log('Listening ...');
  io.listen(3001).on('connect', dockerBackend.addClient.bind(dockerBackend));
}

main();
