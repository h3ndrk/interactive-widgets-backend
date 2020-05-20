const { program } = require('commander');
const getPages = require('./pages');
const ApiServer = require('./ApiServer');
const DockerOperator = require('./DockerOperator');

async function main() {
  let pagesDirectory;
  let monitorWriteDirectory;
  program.arguments('<pages> <monitor-write>').action((pagesValue, monitorWriteValue) => {
    pagesDirectory = pagesValue;
    monitorWriteDirectory = monitorWriteValue;
  });
  program.parse(process.argv);

  const pages = await getPages(pagesDirectory);

  const dockerOperator = new DockerOperator(pages, monitorWriteDirectory);
  await dockerOperator.cleanup();
  await dockerOperator.buildImages();

  const apiServer = new ApiServer(pages);

  apiServer.onStartPage = dockerOperator.startPage.bind(dockerOperator);
  apiServer.onStopPage = dockerOperator.stopPage.bind(dockerOperator);
  apiServer.onButtonClick = dockerOperator.buttonClick.bind(dockerOperator);

  dockerOperator.sendButtonOutput = apiServer.sendButtonOutput.bind(apiServer);

  console.log('Listening at http://localhost:8080/ ...');
  apiServer.listen(8080);
}

main();
