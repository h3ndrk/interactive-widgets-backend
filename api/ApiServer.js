const path = require('path');
const url = require('url');
const EventEmitter = require('events');
const express = require('express');
const SseStream = require('ssestream');
const hashFromPageUuidAndUrl = require('./hashFromPageUuidAndUrl');

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
        data: this.pages,
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
      const requestUrl = new url.URL(path.join(path.sep, path.relative('/page', request.url)), 'http://localhost/');
      const pageUrl = requestUrl.pathname;
      if (!this.pages[pageUrl]) {
        console.log(`${request.url} -> 404`);
        response.sendStatus(404);
        return;
      }
      if (!requestUrl.searchParams.get('pageUuid')) {
        console.log(`${request.url} -> 400`);
        response.sendStatus(400);
        return;
      }
      const stream = new SseStream(request);
      const pageUuid = requestUrl.searchParams.get('pageUuid');
      const pageHash = hashFromPageUuidAndUrl(pageUuid, pageUrl);
      if (!this.pageStreams[pageHash])
        this.emit('instantiate', pageUuid, pageUrl);
      this.pageStreams = {
        ...this.pageStreams,
        [pageHash]: [...(this.pageStreams[pageHash] || []), stream],
      };
      stream.pipe(response);
      stream.write({
        event: 'widgets',
        data: this.pages[pageUrl].widgets,
      });
      response.on('close', () => {
        this.pageStreams = Object.keys(this.pageStreams).reduce((pageStreams, currentPageHash) => ({
          ...pageStreams,
          ...(currentPageHash === pageHash && this.pageStreams[currentPageHash].length === 1 ? {} : { [currentPageHash]: this.pageStreams[currentPageHash].filter(pageStream => pageStream !== stream) }),
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
        data: pages,
      });
    }
    for (const pageHash of Object.keys(this.pageStreams)) {
      for (const pageStream of this.pageStreams[pageHash]) {
        pageStream.write({
          event: 'pages',
          data: pages,
        });
      }
    }
  }
}

module.exports = ApiServer;
