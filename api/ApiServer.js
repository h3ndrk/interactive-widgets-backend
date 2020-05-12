const path = require('path');
const url = require('url');
const crypto = require('crypto');
const EventEmitter = require('events');
const express = require('express');
const SseStream = require('ssestream');

class ApiServer extends EventEmitter {
  constructor() {
    super();
    this.pages = {};
    this.pagesStreams = [];
    this.pageStreams = {};
    this.app = express();
    this.app.disable('x-powered-by');
    this.app.get('/pages', (request, response) => {
      // TODO: test that client stream is added to pagesStream (and removed if disconnected)
      // TODO: test that server emits 'pages' event on connect
      const stream = new SseStream(request);
      this.pagesStreams = [...this.pagesStreams, stream];
      stream.pipe(response);
      stream.write({
        event: 'pages',
        data: Object.keys(this.pages),
      });
      response.on('close', () => {
        this.pagesStreams = this.pagesStreams.filter(pagesStream => pagesStream !== stream);
        stream.unpipe(response);
      });
    });
    this.app.get(['/page', '/page/*'], (request, response) => {
      // TODO: test that client stream is added to pagesStream (and removed if disconnected)
      // TODO: test 404 and 400 (negatives and positives)
      // TODO: test that server emits 'widgets' event on connect
      // TODO: test that hashes are unique
      const requestUrl = new url.URL(request.url, 'http://localhost/');
      const pageUrl = path.join(path.sep, path.relative('/page', requestUrl.pathname));
      if (!this.pages[pageUrl]) {
        console.log(`${request.url} (${pageUrl}) -> 404`);
        response.sendStatus(404);
        return;
      }
      if (!requestUrl.searchParams.get('pageUuid')) {
        console.log(`${request.url} (${pageUrl}) -> 400`);
        response.sendStatus(400);
        return;
      }
      const stream = new SseStream(request);
      const pageUuid = requestUrl.searchParams.get('pageUuid');
      const pageHash = crypto.createHash('sha256').update(`${pageUuid} ${pageUrl}`).digest('hex');
      const pageStream = {
        stream: stream,
        uuid: pageUuid,
        url: pageUrl,
      };
      if (!this.pageStreams[pageHash])
        this.emit('instantiate', pageUuid, pageUrl);
      this.pageStreams = {
        ...this.pageStreams,
        [pageHash]: [...(this.pageStreams[pageHash] || []), pageStream],
      };
      stream.pipe(response);
      stream.write({
        event: 'widgets',
        data: this.convertWidgets(pageUuid, pageUrl),
      });
      response.on('close', () => {
        this.pageStreams = Object.keys(this.pageStreams).reduce((pageStreams, currentPageHash) => ({
          ...pageStreams,
          ...(currentPageHash === pageHash && this.pageStreams[currentPageHash].length === 1 ? {} : {
            [currentPageHash]: this.pageStreams[currentPageHash].filter(currentPageStream => currentPageStream !== pageStream),
          }),
        }), {});
        stream.unpipe(response);
        if (!this.pageStreams[pageHash])
          this.emit('teardown', pageUuid, pageUrl);
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
  setPages(pages) {
    // TODO: test setting pages, while connected to /pages and /page/...
    this.pages = pages;
    for (const pagesStream of this.pagesStreams) {
      pagesStream.write({
        event: 'pages',
        data: Object.keys(pages),
      });
    }
    for (const pageHash of Object.keys(this.pageStreams)) {
      for (const pageStream of this.pageStreams[pageHash]) {
        pageStream.stream.write({
          event: 'pages',
          data: Object.keys(pages),
        });
        pageStream.stream.write({
          event: 'widgets',
          data: this.convertWidgets(pageStream.uuid, pageStream.url),
        });
      }
    }
  }
  convertWidgets(pageUuid, pageUrl) {
    return this.pages[pageUrl].widgets.map((widget, index) => {
      switch (widget.type) {
        case 'text':
          return {
            ...widget,
            hash: crypto.createHash('sha256').update(`${pageUuid} ${pageUrl} ${widget.file} ${index}`).digest('hex'),
          };
        case 'image':
          return {
            ...widget,
            hash: crypto.createHash('sha256').update(`${pageUuid} ${pageUrl} ${widget.file} ${index}`).digest('hex'),
          };
        case 'button':
          return {
            ...widget,
            hash: crypto.createHash('sha256').update(`${pageUuid} ${pageUrl} ${widget.label} ${widget.command} ${index}`).digest('hex'),
          };
        case 'editor':
          return {
            ...widget,
            hash: crypto.createHash('sha256').update(`${pageUuid} ${pageUrl} ${widget.file} ${index}`).digest('hex'),
          };
        case 'terminal':
          return {
            ...widget,
            hash: crypto.createHash('sha256').update(`${pageUuid} ${pageUrl} ${widget.workingDirectory} ${index}`).digest('hex'),
          };
        default:
          return widget;
      }
    });
  }
}

module.exports = ApiServer;
