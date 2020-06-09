import React, { useEffect, useState, useRef } from 'react';
import { useLocation } from 'react-router-dom';
import MarkdownWidget from './widget/MarkdownWidget';
import TextWidget from './widget/TextWidget';
import ImageWidget from './widget/ImageWidget';
import ButtonWidget from './widget/ButtonWidget';
import EditorWidget from './widget/EditorWidget';
import TerminalWidget from './widget/TerminalWidget';

export default function InteractivePage(props) {
  const [widgets, setWidgets] = useState(props.widgets);
  const location = useLocation();
  const roomID = location.hash.substr(1);
  const webSocket = useRef(null);
  const terminalWidgets = useRef([]);
  
  useEffect(() => {
    if (widgets) {
      terminalWidgets.current = terminalWidgets.current.slice(0, widgets.length);
    }
  }, [widgets]);
  
  useEffect(() => {
    console.log('Connecting to WebSocket ...');
    webSocket.current = new WebSocket(`${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}/api/page/attach?page_url=${encodeURIComponent(props.page.url)}&room_id=${encodeURIComponent(roomID)}`);
    webSocket.current.onclose = event => {
      alert('Connection to server lost. The page might not work anymore. Please reload to reconnect.');
    };
    webSocket.current.onmessage = event => {
      try {
        const message = JSON.parse(event.data);

        if (!message.widgetIndex) {
          console.warn('Ignore message: Missing widgetIndex:', event.data);
          return;
        }

        if (typeof message.widgetIndex !== 'number') {
          console.warn('Ignore message: widgetIndex is not a number:', event.data);
          return;
        }

        if (!message.data) {
          console.warn('Ignore message: Missing data:', event.data);
          return;
        }

        if (message.widgetIndex < 0 || message.widgetIndex >= props.widgets.length) {
          console.warn('Ignore message: widgetIndex is out of bounds:', event.data, props.widgets);
          return;
        }
        
        // since setState functions may be called multiple times, we need to process stateful messages outside of the function
        // https://reactjs.org/docs/strict-mode.html#detecting-unexpected-side-effects
        if (props.widgets[message.widgetIndex].type === 'terminal' && message.data.data) {
          terminalWidgets.current[message.widgetIndex].write(atob(message.data.data));
          return;
        }

        const patchedMonitorWriteWidget = (widget, message, decodeBase64) => {
          if (message.data.type !== undefined) {
            switch (message.data.type) {
              case "jsonError":
                if (message.data.errorReason !== undefined) {
                  return { ...widget, contents: null, error: `JSON Error: ${message.data.errorReason}` };
                }
                break;
              case "argumentError":
                if (message.data.expectedCount !== undefined && message.data.gotCount !== undefined && message.data.gotArguments !== undefined) {
                  return { ...widget, contents: null, error: `monitor-write process called with wrong arguments: Expected ${message.data.expectedCount}, got ${message.data.gotCount}: [${message.data.gotArguments.map(argument => `"${argument}"`).join(", ")}]` };
                }
                break;
              case "stdinReadError":
                if (message.data.errorReason !== undefined) {
                  return { ...widget, contents: null, error: `Error while reading stdin: ${message.data.errorReason}` };
                }
                break;
              case "removalError":
                if (message.data.path !== undefined && message.data.errorReason !== undefined) {
                  return { ...widget, contents: null, error: `Error while removing the watched file "${message.data.path}": ${message.data.errorReason}` };
                }
                break;
              case "contents":
                if (message.data.contents !== undefined) {
                  return { ...widget, contents: decodeBase64 ? atob(message.data.contents) : message.data.contents, error: null };
                }
                break;
              case "openError":
                if (message.data.path !== undefined && message.data.errorReason !== undefined) {
                  return { ...widget, contents: null, error: `Error while opening the file "${message.data.path}" for reading: ${message.data.errorReason}` };
                }
                break;
              case "readAndDecodeError":
                if (message.data.errorReason !== undefined) {
                  return { ...widget, contents: null, error: `Error while reading and decoding Base64 data: ${message.data.errorReason}` };
                }
                break;
              case "createWatcherError":
                if (message.data.errorReason !== undefined) {
                  return { ...widget, contents: null, error: `Error while creating a file watcher: ${message.data.errorReason}` };
                }
                break;
              case "addWatcherError":
                if (message.data.path !== undefined && message.data.errorReason !== undefined) {
                  return { ...widget, contents: null, error: `Error while creating a file watcher for "${message.data.path}": ${message.data.errorReason}` };
                }
                break;
              case "watchError":
                if (message.data.path !== undefined && message.data.errorReason !== undefined) {
                  return { ...widget, contents: null, error: `Error while watching the file "${message.data.path}": ${message.data.errorReason}` };
                }
                break;
              case "createError":
                if (message.data.path !== undefined && message.data.errorReason !== undefined) {
                  return { ...widget, contents: null, error: `Error while creating the file "${message.data.path}": ${message.data.errorReason}` };
                }
                break;
              case "decodeError":
                if (message.data.errorReason !== undefined) {
                  return { ...widget, contents: null, error: `Error while decoding Base64 data: ${message.data.errorReason}` };
                }
                break;
              case "writeError":
                if (message.data.errorReason !== undefined) {
                  return { ...widget, contents: null, error: `Error while writing Base64 data to a file: ${message.data.errorReason}` };
                }
                break;
              default:
            }
          }
        };

        setWidgets(widgets => widgets.map((widget, widgetIndex) => {
          if (widgetIndex !== message.widgetIndex) {
            return widget;
          }

          switch (widget.type) {
            case 'text': {
              const patchedWidget = patchedMonitorWriteWidget(widget, message, true);
              if (patchedWidget !== undefined) {
                return patchedWidget;
              }

              console.warn('Ignore message: TextWidget does not have handler implemented', event.data);
              return widget;
            }
            case 'image': {
              const patchedWidget = patchedMonitorWriteWidget(widget, message, false);
              if (patchedWidget !== undefined) {
                return patchedWidget;
              }

              console.warn('Ignore message: ImageWidget does not have handler implemented', event.data);
              return widget;
            }
            case 'button': {
              if (message.data.clear === true) {
                return { ...widget, outputs: [] };
              } else if (message.data.origin && ['stdout', 'stderr'].includes(message.data.origin) && message.data.data) {
                return {
                  ...widget,
                  outputs: [
                    ...widget.outputs,
                    {
                      origin: message.data.origin,
                      data: atob(message.data.data),
                    }
                  ],
                };
              }

              console.warn('Ignore message: ButtonWidget does not have handler implemented', event.data);
              return widget;
            }
            case 'editor': {
              const patchedWidget = patchedMonitorWriteWidget(widget, message, true);
              if (patchedWidget !== undefined) {
                return patchedWidget;
              }

              console.warn('Ignore message: EditorWidget does not have handler implemented', event.data);
              return widget;
            }
            case 'terminal': {
              console.warn('Bug: Messages for terminal widgets should have been handeled earlier. Ignoring message.');
              return widget;
            }
            default:
              console.warn('Ignore message: Widget with widgetIndex does not have handler implemented', event.data, widgets);
              return widget;
          }
        }));
      } catch (e) {
        console.warn('Got error while processing message:', e, event.data);
      }
    };
    
    return () => {
      if (webSocket.current !== null) {
        console.log('Disconnecting from WebSocket ...');
        webSocket.current.onclose = () => { };
        webSocket.current.close();
      }
    };
  }, [props.page.url, roomID, props.widgets]);

  return (
    <>
      {widgets !== null && widgets.map((widget, i) => {
        switch (widget.type) {
          case 'markdown':
            return (
              <MarkdownWidget
                key={`${props.page.url}/${i}`}
                widget={widget}
                pageURL={props.page.url}
              />
            );
          case 'text':
            return (
              <TextWidget
                key={`${props.page.url}/${i}`}
                widget={widget}
              />
            );
          case 'image':
            return (
              <ImageWidget
                key={`${props.page.url}/${i}`}
                widget={widget}
              />
            );
          case 'button':
            return (
              <ButtonWidget
                key={`${props.page.url}/${i}`}
                widget={widget}
                onClick={() => {
                  if (webSocket.current !== null) {
                    console.log('click');
                    webSocket.current.send(JSON.stringify({
                      widgetIndex: i,
                      data: {
                        click: true,
                      },
                    }));
                  }
                }}
              />
            );
          case 'editor':
            return (
              <EditorWidget
                key={`${props.page.url}/${i}`}
                widget={widget}
                onCreateOrEmpty={() => {
                  if (webSocket.current !== null) {
                    webSocket.current.send(JSON.stringify({
                      widgetIndex: i,
                      data: {
                        type: "contents",
                        contents: "",
                      },
                    }));
                  }
                }}
                onChange={value => {
                  if (webSocket.current !== null) {
                    webSocket.current.send(JSON.stringify({
                      widgetIndex: i,
                      data: {
                        type: "contents",
                        contents: btoa(value),
                      },
                    }));
                  }
                }}
                onDelete={() => {
                  if (webSocket.current !== null) {
                    webSocket.current.send(JSON.stringify({
                      widgetIndex: i,
                      data: {
                        type: "removal",
                      },
                    }));
                  }
                }}
              />
            );
          case 'terminal':
            return (
              <TerminalWidget
                key={`${props.page.url}/${i}`}
                ref={element => terminalWidgets.current[i] = element}
                widget={widget}
                onData={data => {
                  if (webSocket.current !== null) {
                    webSocket.current.send(JSON.stringify({
                      widgetIndex: i,
                      data: {
                        data: btoa(data),
                      },
                    }));
                  }
                }}
              />
            );
          default:
            return (<div key={`${props.page.url}/${i}`} className="centered">Widget "{widget.type}" not implemented.</div>);
        }
      })}
    </>
  );
}
