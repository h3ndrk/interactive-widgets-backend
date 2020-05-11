const crypto = require('crypto');

module.exports = (pageUuid, url) => {
  return crypto.createHash('sha256').update(`${pageUuid} ${url}`).digest('hex');
};
