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
  const [disconnected, setDisconnected] = useState(false);
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

        setWidgets(widgets => widgets.map((widget, widgetIndex) => {
          if (widgetIndex !== message.widgetIndex) {
            return widget;
          }

          switch (widget.type) {
            case 'text': {
              if (message.data.contents) {
                return { ...widget, contents: atob(message.data.contents), error: null };
              } else if (message.data.error) {
                return { ...widget, contents: null, error: atob(message.data.error) };
              }

              console.warn('Ignore message: TextWidget does not have handler implemented', event.data);
              return widget;
            }
            case 'image': {
              if (message.data.contents) {
                return { ...widget, contents: message.data.contents, error: null };
              } else if (message.data.error) {
                return { ...widget, contents: null, error: atob(message.data.error) };
              }

              console.warn('Ignore message: ImageWidget does not have handler implemented', event.data);
              return widget;
            }
            case 'button': {
              if (message.data.clear === true) {
                return { ...widget, outputs: [] };
              } else if (message.data.origin && ['stdout', 'stderr'].includes(message.data.origin) && message.data.data) {
                return { ...widget, outputs: [...widget.outputs, atob(message.data.data)] };
              }

              console.warn('Ignore message: ButtonWidget does not have handler implemented', event.data);
              return widget;
            }
            case 'editor': {
              if (message.data.contents) {
                return { ...widget, contents: atob(message.data.contents), error: null };
              } else if (message.data.error) {
                return { ...widget, contents: null, error: atob(message.data.error) };
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
      if (this.webSocket !== null) {
        console.log('Disconnecting from WebSocket ...');
        this.webSocket.close();
      }
    };
  }, [props.page.url, roomID, props.widgets]);

  return (
    <>
      {widgets !== null && widgets.map((widget, i) => {
        switch (widget.type) {
          case 'markdown':
            return (<MarkdownWidget key={`${props.page.url}/${i}`} widget={widget} />);
          case 'text':
            return (<TextWidget key={`${props.page.url}/${i}`} widget={widget} />);
          case 'image':
            return (<ImageWidget key={`${props.page.url}/${i}`} widget={widget} />);
          case 'button':
            return (<ButtonWidget key={`${props.page.url}/${i}`} widget={widget} />);
          case 'editor':
            return (<EditorWidget key={`${props.page.url}/${i}`} widget={widget} />);
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
