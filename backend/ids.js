module.exports = {
  urlAndUuidToPageId: (url, uuid) => {
    if (!url || typeof url !== 'string' || url.length === 0 || url[0] !== '/')
      throw new Error(`Malformed URL: ${url}`);
    if (!uuid || typeof uuid !== 'string' || uuid.length !== 36 || !uuid.match(/^[0-9A-F]{8}-[0-9A-F]{4}-4[0-9A-F]{3}-[89AB][0-9A-F]{3}-[0-9A-F]{12}$/i))
      throw new Error(`Malformed UUID: ${uuid}`);
    if (url.slice(-1) === '/')
      url = url.substr(0, url.length - 1);
    return `${url}/${uuid}`;
  },
  pageIdToUrlAndUuid: pageId => {
    if (!pageId || typeof pageId !== 'string' || pageId.length === 0)
      throw new Error(`Malformed PageID: ${pageId}`);
    const matches = pageId.match(/^(\/.+)*\/([0-9A-F]{8}-[0-9A-F]{4}-4[0-9A-F]{3}-[89AB][0-9A-F]{3}-[0-9A-F]{12})$/i);
    if (!matches)
      throw new Error(`Malformed PageID: ${pageId}`);
    return {
      url: matches[1],
      uuid: matches[2],
    };
  },
  urlAndUuidAndWidgetIndexToWidgetId: (url, uuid, widgetIndex) => {
    if (!url || typeof url !== 'string' || url.length === 0 || url[0] !== '/')
      throw new Error(`Malformed URL: ${url}`);
    if (!uuid || typeof uuid !== 'string' || uuid.length !== 36 || !uuid.match(/^[0-9A-F]{8}-[0-9A-F]{4}-4[0-9A-F]{3}-[89AB][0-9A-F]{3}-[0-9A-F]{12}$/i))
      throw new Error(`Malformed UUID: ${uuid}`);
      if (!widgetIndex || typeof widgetIndex !== 'number' || isNaN(widgetIndex))
      throw new Error(`Malformed WidgetIndex: ${widgetIndex}`);
    if (url.slice(-1) === '/')
      url = url.substr(0, url.length - 1);
    return `${url}/${uuid}/${widgetIndex}`;
  },
  widgetIdToUrlAndUuidAndWidgetIndex: widgetId => {
    if (!widgetId || typeof widgetId !== 'string' || widgetId.length === 0)
      throw new Error(`Malformed WidgetID: ${widgetId}`);
    const matches = widgetId.match(/^(\/.+)*\/([0-9A-F]{8}-[0-9A-F]{4}-4[0-9A-F]{3}-[89AB][0-9A-F]{3}-[0-9A-F]{12})\/(\d+)$/i);
    if (!matches)
      throw new Error(`Malformed WidgetID: ${widgetId}`);
    return {
      url: matches[1],
      uuid: matches[2],
      widgetIndex: parseInt(matches[3]),
    };
  },
  idToEncodedId: id => {
    if (!id || typeof id !== 'string' || id.length === 0)
      throw new Error(`Malformed ID: ${id}`);
    return Buffer.from(id).toString('hex');
  },
  encodedIdToId: encodedId => {
    if (!encodedId || typeof encodedId !== 'string' || encodedId.length === 0)
      throw new Error(`Malformed ID: ${encodedId}`);
    return Buffer.from(encodedId, 'hex').toString();
  },
};
