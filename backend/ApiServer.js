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
    this.app = express();
    this.app.disable('x-powered-by');
    this.app.use(express.json());
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
    this.app.get(['/page/widget/:widgetIndex/button-click/:uuid', '/page/*/widget/:widgetIndex/button-click/:uuid'], (request, response) => {
      const requestUrl = new url.URL(request.url, 'http://localhost/');
      const trimmedPath = path.join(path.sep, path.relative('/page', requestUrl.pathname));
      const pageUrl = path.dirname(path.dirname(path.dirname(path.dirname(trimmedPath))));
      if (!this.pages[pageUrl]) {
        console.log(`${request.url} -> 404 (page not found)`);
        response.sendStatus(404);
        return;
      }
      if (!request.params['widgetIndex']) {
        console.log(`${request.url} -> 400 (widget index missing)`);
        response.sendStatus(400);
        return;
      }
      // https://gist.github.com/johnelliott/cf77003f72f889abbc3f32785fa3df8d
      if (isNaN(request.params['widgetIndex'])) {
        console.log(`${request.url} -> 400 (malformed widget index)`);
        response.sendStatus(400);
        return;
      }
      const widgetIndex = parseInt(request.params['widgetIndex']);
      if (this.pages[pageUrl].widgets[widgetIndex].type !== 'button') {
        console.log(`${request.url} -> 400 (wrong widget type)`);
        response.sendStatus(400);
        return;
      }
      if (!request.params['uuid']) {
        console.log(`${request.url} -> 400 (uuid missing)`);
        response.sendStatus(400);
        return;
      }
      const uuid = request.params['uuid'];
      // https://gist.github.com/johnelliott/cf77003f72f889abbc3f32785fa3df8d
      if (!uuid.match(/^[0-9A-F]{8}-[0-9A-F]{4}-4[0-9A-F]{3}-[89AB][0-9A-F]{3}-[0-9A-F]{12}$/i)) {
        console.log(`${request.url} -> 400 (malformed uuid)`);
        response.sendStatus(400);
        return;
      }
      this.emit('buttonClick', path.join(pageUrl, uuid, widgetIndex.toString()));
      response.sendStatus(200);
    });
    this.app.post(['/page/widget/:widgetIndex/editor-contents/:uuid', '/page/*/widget/:widgetIndex/editor-contents/:uuid'], (request, response) => {
      const requestUrl = new url.URL(request.url, 'http://localhost/');
      const trimmedPath = path.join(path.sep, path.relative('/page', requestUrl.pathname));
      const pageUrl = path.dirname(path.dirname(path.dirname(path.dirname(trimmedPath))));
      if (!this.pages[pageUrl]) {
        console.log(`${request.url} -> 404 (page not found)`);
        response.sendStatus(404);
        return;
      }
      if (!request.params['widgetIndex']) {
        console.log(`${request.url} -> 400 (widget index missing)`);
        response.sendStatus(400);
        return;
      }
      // https://gist.github.com/johnelliott/cf77003f72f889abbc3f32785fa3df8d
      if (isNaN(request.params['widgetIndex'])) {
        console.log(`${request.url} -> 400 (malformed widget index)`);
        response.sendStatus(400);
        return;
      }
      const widgetIndex = parseInt(request.params['widgetIndex']);
      if (this.pages[pageUrl].widgets[widgetIndex].type !== 'editor') {
        console.log(`${request.url} -> 400 (wrong widget type)`);
        response.sendStatus(400);
        return;
      }
      if (!request.params['uuid']) {
        console.log(`${request.url} -> 400 (uuid missing)`);
        response.sendStatus(400);
        return;
      }
      const uuid = request.params['uuid'];
      // https://gist.github.com/johnelliott/cf77003f72f889abbc3f32785fa3df8d
      if (!uuid.match(/^[0-9A-F]{8}-[0-9A-F]{4}-4[0-9A-F]{3}-[89AB][0-9A-F]{3}-[0-9A-F]{12}$/i)) {
        console.log(`${request.url} -> 400 (malformed uuid)`);
        response.sendStatus(400);
        return;
      }
      this.emit('editorContents', path.join(pageUrl, uuid, widgetIndex.toString()), request.body);
      response.sendStatus(200);
    });
    this.app.post(['/page/widget/:widgetIndex/terminal-input/:uuid', '/page/*/widget/:widgetIndex/terminal-input/:uuid'], (request, response) => {
      const requestUrl = new url.URL(request.url, 'http://localhost/');
      const trimmedPath = path.join(path.sep, path.relative('/page', requestUrl.pathname));
      const pageUrl = path.dirname(path.dirname(path.dirname(path.dirname(trimmedPath))));
      if (!this.pages[pageUrl]) {
        console.log(`${request.url} -> 404 (page not found)`);
        response.sendStatus(404);
        return;
      }
      if (!request.params['widgetIndex']) {
        console.log(`${request.url} -> 400 (widget index missing)`);
        response.sendStatus(400);
        return;
      }
      // https://gist.github.com/johnelliott/cf77003f72f889abbc3f32785fa3df8d
      if (isNaN(request.params['widgetIndex'])) {
        console.log(`${request.url} -> 400 (malformed widget index)`);
        response.sendStatus(400);
        return;
      }
      const widgetIndex = parseInt(request.params['widgetIndex']);
      if (this.pages[pageUrl].widgets[widgetIndex].type !== 'terminal') {
        console.log(`${request.url} -> 400 (wrong widget type)`);
        response.sendStatus(400);
        return;
      }
      if (!request.params['uuid']) {
        console.log(`${request.url} -> 400 (uuid missing)`);
        response.sendStatus(400);
        return;
      }
      const uuid = request.params['uuid'];
      // https://gist.github.com/johnelliott/cf77003f72f889abbc3f32785fa3df8d
      if (!uuid.match(/^[0-9A-F]{8}-[0-9A-F]{4}-4[0-9A-F]{3}-[89AB][0-9A-F]{3}-[0-9A-F]{12}$/i)) {
        console.log(`${request.url} -> 400 (malformed uuid)`);
        response.sendStatus(400);
        return;
      }
      this.emit('terminalInput', path.join(pageUrl, uuid, widgetIndex.toString()), request.body);
      response.sendStatus(200);
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
  sendTextContents(widgetId, contents) {
    const widgetIndex = parseInt(path.basename(widgetId));
    const pageId = path.dirname(widgetId);
    const pageUrl = path.dirname(pageId);
    if (!this.streams[pageId] || this.pages[pageUrl].widgets[widgetIndex].type !== 'text') {
      console.log(`Dropping text contents for ${widgetId} ... (widget not existing or wrong type)`);
      return;
    }
    this.streams[pageId].forEach(stream => {
      stream.write({
        event: 'textContents',
        data: {
          widgetId: widgetIndex,
          contents: contents,
        },
      })
    });
  }
  sendImageData(widgetId, data) {
    const widgetIndex = parseInt(path.basename(widgetId));
    const pageId = path.dirname(widgetId);
    const pageUrl = path.dirname(pageId);
    if (!this.streams[pageId] || this.pages[pageUrl].widgets[widgetIndex].type !== 'image') {
      console.log(`Dropping image data for ${widgetId} ... (widget not existing or wrong type)`);
      return;
    }
    this.streams[pageId].forEach(stream => {
      stream.write({
        event: 'imageData',
        data: {
          widgetId: widgetIndex,
          data: data,
        },
      })
    });
  }
  sendButtonOutput(widgetId, output) {
    const widgetIndex = parseInt(path.basename(widgetId));
    const pageId = path.dirname(widgetId);
    const pageUrl = path.dirname(pageId);
    if (!this.streams[pageId] || this.pages[pageUrl].widgets[widgetIndex].type !== 'button') {
      console.log(`Dropping button output for ${widgetId} ... (widget not existing or wrong type)`);
      return;
    }
    this.streams[pageId].forEach(stream => {
      stream.write({
        event: 'buttonOutput',
        data: {
          widgetId: widgetIndex,
          output: output,
        },
      })
    });
  }
  sendEditorContents(widgetId, contents) {
    const widgetIndex = parseInt(path.basename(widgetId));
    const pageId = path.dirname(widgetId);
    const pageUrl = path.dirname(pageId);
    if (!this.streams[pageId] || this.pages[pageUrl].widgets[widgetIndex].type !== 'editor') {
      console.log(`Dropping editor contents for ${widgetId} ... (widget not existing or wrong type)`);
      return;
    }
    this.streams[pageId].forEach(stream => {
      stream.write({
        event: 'editorContents',
        data: {
          widgetId: widgetIndex,
          contents: contents,
        },
      })
    });}
  sendTerminalOutput(widgetId, output) {
    const widgetIndex = parseInt(path.basename(widgetId));
    const pageId = path.dirname(widgetId);
    const pageUrl = path.dirname(pageId);
    if (!this.streams[pageId] || this.pages[pageUrl].widgets[widgetIndex].type !== 'terminal') {
      console.log(`Dropping terminal output for ${widgetId} ... (widget not existing or wrong type)`);
      return;
    }
    this.streams[pageId].forEach(stream => {
      stream.write({
        event: 'terminalOutput',
        data: {
          widgetId: widgetIndex,
          output: output,
        },
      })
    });
  }
}

module.exports = ApiServer;
