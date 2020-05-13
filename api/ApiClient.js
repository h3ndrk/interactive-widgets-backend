class ApiClient extends EventTarget {
  constructor(endpoint) {
    super();
    this.endpoint = endpoint;
    this.pages = {};
    this.widgets = [];
    this.pagesReceived = false;
    this.url = null;
    this.uuid = null;
    this.eventSource = null;
    fetch(`${this.endpoint}/pages`)
      .then(response => {
        if (response.status >= 400)
          throw new Error(`Got ${response.status}`);
        return response.json();
      })
      .then(pages => {
        this.pages = pages;
        this.pagesReceived = true;
        this.dispatchEvent(new CustomEvent('pagesChanged', { detail: pages }));
        if (this.url === null || this.uuid === null)
          return;
        this._connectWidgets();
      });
  }
  setUrlAndUuid(url, uuid) {
    this.url = url;
    this.uuid = uuid;
    if (!this.pagesReceived)
      return;
    this._connectWidgets();
  }
  _connectWidgets() {
    fetch(`${this.endpoint}/page${this.url === '/' ? '' : this.url}/widgets`)
      .then(response => {
        if (response.status >= 400)
          throw new Error(`Got ${response.status}`);
        return response.json();
      })
      .then(widgets => {
        this.widgets = widgets;
        this.dispatchEvent(new CustomEvent('widgetsChanged', { detail: widgets }));
        if (this.eventSource)
          this.eventSource.close();
        this.eventSource = new EventSource(`${this.endpoint}/page${this.url === '/' ? '' : this.url}/updates/${this.uuid}`);
        this.eventSource.addEventListener('textContents', event => {
          this.dispatchEvent(new CustomEvent('textContents', { detail: event }));
        });
        this.eventSource.addEventListener('imageData', event => {
          this.dispatchEvent(new CustomEvent('imageData', { detail: event }));
        });
        this.eventSource.addEventListener('buttonOutput', event => {
          this.dispatchEvent(new CustomEvent('buttonOutput', { detail: event }));
        });
        this.eventSource.addEventListener('editorContents', event => {
          this.dispatchEvent(new CustomEvent('editorContents', { detail: event }));
        });
        this.eventSource.addEventListener('terminalOutput', event => {
          this.dispatchEvent(new CustomEvent('terminalOutput', { detail: event }));
        });
      });
  }
  sendButtonClick(widgetIndex) {
    if (isNaN(widgetIndex))
      throw new Error(`Widget index is NaN: ${widgetIndex}`);
    if (this.url === null || this.uuid === null)
      return;
    fetch(`${this.endpoint}/page${this.url === '/' ? '' : this.url}/widget/${widgetIndex}/button-click/${this.uuid}`)
      .then(response => {
        if (response.status >= 400)
          throw new Error(`Got ${response.status}`);
      });
  }
  sendEditorContents(widgetIndex, contents) {
    if (isNaN(widgetIndex))
      throw new Error(`Widget index is NaN: ${widgetIndex}`);
    if (this.url === null || this.uuid === null)
      return;
    fetch(`${this.endpoint}/page${this.url === '/' ? '' : this.url}/widget/${widgetIndex}/editor-contents/${this.uuid}`, { method: 'POST', body: JSON.stringify(contents) })
      .then(response => {
        if (response.status >= 400)
          throw new Error(`Got ${response.status}`);
      });
  }
  sendTerminalInput(widgetIndex, input) {
    if (isNaN(widgetIndex))
      throw new Error(`Widget index is NaN: ${widgetIndex}`);
    if (this.url === null || this.uuid === null)
      return;
    fetch(`${this.endpoint}/page${this.url === '/' ? '' : this.url}/widget/${widgetIndex}/terminal-input/${this.uuid}`, { method: 'POST', body: JSON.stringify(input) })
      .then(response => {
        if (response.status >= 400)
          throw new Error(`Got ${response.status}`);
      });
  }
}

defineEventAttribute(ApiClient.prototype, 'pagesChanged');
defineEventAttribute(ApiClient.prototype, 'widgetsChanged');
defineEventAttribute(ApiClient.prototype, 'textContents');
defineEventAttribute(ApiClient.prototype, 'imageData');
defineEventAttribute(ApiClient.prototype, 'buttonOutput');
defineEventAttribute(ApiClient.prototype, 'editorContents');
defineEventAttribute(ApiClient.prototype, 'terminalOutput');

module.exports = ApiClient;
