const { program } = require('commander');
const Docker = require('dockerode');
const getPages = require('./pages');
const ApiServer = require('./ApiServer');
const DockerOperator = require('./DockerOperator');

async function main() {
  let pagesDirectory;
  program.arguments('<pages>').action(pagesValue => {
    pagesDirectory = pagesValue;
  });
  program.parse(process.argv);

  const pages = await getPages(pagesDirectory);

  const dockerOperator = new DockerOperator(pages);
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
