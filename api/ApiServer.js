const path = require('path');
const url = require('url');
const crypto = require('crypto');
const EventEmitter = require('events');
const express = require('express');
const SseStream = require('ssestream');

class ApiServer extends EventEmitter {
  constructor(pages) {
    super();
    this.pages = pages;
    this.streams = {};
    this.widgetsToStreams = {};
    this.pagesStreams = [];
    this.pageStreams = {};
    this.activeWidgets = {};
    this.app = express();
    this.app.disable('x-powered-by');
    this.app.get('/pages', (_, response) => {
      // TODO: test that server responds with 'pages'
      response.json(Object.keys(this.pages));
    });
    this.app.get(['/page/widgets', '/page/*/widgets'], (request, response) => {
      const requestUrl = new url.URL(request.url, 'http://localhost/');
      const trimmedPath = path.join(path.sep, path.relative('/page', requestUrl.pathname));
      const pageUrl = path.dirname(trimmedPath);
      if (!this.pages[pageUrl]) {
        console.log(`${request.url} -> 404 (page not found)`);
        response.sendStatus(404);
        return;
      }
      response.json(this.pages[pageUrl].widgets);
    });
    this.app.get(['/page/updates/:uuid', '/page/*/updates/:uuid'], (request, response) => {
      // TODO: test that client stream is added to streams (and removed if disconnected)
      // TODO: test 404 and 400 (negatives and positives)
      const requestUrl = new url.URL(request.url, 'http://localhost/');
      const trimmedPath = path.join(path.sep, path.relative('/page', requestUrl.pathname));
      const pageUrl = path.dirname(path.dirname(trimmedPath));
      if (!this.pages[pageUrl]) {
        console.log(`${request.url} -> 404 (page not found)`);
        response.sendStatus(404);
        return;
      }
      if (!request.params['uuid']) {
        console.log(`${request.url} (${pageUrl}) -> 400 (uuid missing)`);
        response.sendStatus(400);
        return;
      }
      const uuid = request.params['uuid'];
      // https://gist.github.com/johnelliott/cf77003f72f889abbc3f32785fa3df8d
      if (!uuid.match(/^[0-9A-F]{8}-[0-9A-F]{4}-4[0-9A-F]{3}-[89AB][0-9A-F]{3}-[0-9A-F]{12}$/i)) {
        console.log(`${request.url} (${pageUrl}) -> 400 (malformed uuid)`);
        response.sendStatus(400);
        return;
      }
      const pageId = path.join(pageUrl, uuid);
      const stream = new SseStream(request);
      stream.pipe(response);
      if (!this.streams[pageId])
        this.emit('instantiate', pageId);
      this.streams = {
        ...this.streams,
        [pageId]: [...(this.streams[pageId] || []), stream],
      };
      response.on('close', () => {
        this.streams = Object.keys(this.streams).reduce((streams, pageId_) => ({
          ...streams,
          ...(pageId_ === pageId && this.streams[pageId_].length === 1 ? {} : {
            [pageId_]: this.streams[pageId_].filter(pageStream => pageStream !== stream),
          }),
        }), {});
        stream.unpipe(response);
        if (!this.streams[pageId])
          this.emit('teardown', pageId);
      });
    });
  }
  listen(port) {
    // TODO: test that port is open (only after calling listen(), not before)
    this.app.listen(port, error => {
      if (error)
        throw error;
    });
  }
  // widgetHash is public
  // pageHash is private
  // widgetHash -> stream
  // emits: 'instantiate' (pageHash)
  // emits: 'teardown' (pageHash)
  sendTextData(widgetHash, data) { }
  sendImageData(widgetHash, data) { }
  // emits: 'executeButton' (pageHash, widgetHash)
  sendButtonOutput(widgetHash, output) { }
  // emits: 'sendEditorData' (pageHash, widgetHash)
  sendEditorData(widgetHash, data) { }
  sendTerminalOutput(widgetHash, output) { }
}

module.exports = ApiServer;
