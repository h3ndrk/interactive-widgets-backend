const { program } = require('commander');
const Docker = require('dockerode');
const getPages = require('./getPages');
const ApiServer = require('api/ApiServer');
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

  const apiServer = new ApiServer(pages);
  const dockerBackend = new DockerBackend(docker, pages, apiServer);
  await dockerBackend.buildPages();
  apiServer.listen(8080);
}

main();
