import React, { useEffect, useState, useRef } from 'react';
import { Link, useLocation } from 'react-router-dom';
import MarkdownWidget from './widgets/MarkdownWidget';
import TextWidget from './widgets/TextWidget';
import ImageWidget from './widgets/ImageWidget';
import ButtonWidget from './widgets/ButtonWidget';
import EditorWidget from './widgets/EditorWidget';
import TerminalWidget from './widgets/TerminalWidget';

function handleMessage(event, widgets, setWidgets, terminalWidgets) {
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
    
    if (message.widgetIndex < 0 || message.widgetIndex >= widgets.length) {
      console.warn('Ignore message: widgetIndex is out of bounds:', event.data, widgets);
      return widgets;
    }
    
    setWidgets(widgets.map((widget, widgetIndex) => {
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
          if (message.data.data) {
            terminalWidgets.current[widgetIndex].write(atob(message.data.data));
            return widget;
          }
          
          console.warn('Ignore message: TerminalWidget does not have handler implemented', event.data);
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
}

export default function Page(props) {
  const [page, setPage] = useState(props.page);
  const [widgets, setWidgets] = useState(null);
  const [disconnected, setDisconnected] = useState(false);
  const location = useLocation();
  const roomID = location.hash.substr(1);
  const webSocket = useRef(null);
  const terminalWidgets = useRef([]);
  
  useEffect(() => {
    if (widgets && page.isInteractive && webSocket.current === null) {
      console.log('Connecting to WebSocket ...');
      webSocket.current = new WebSocket(`${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}/api/page/attach?page_url=${encodeURIComponent(props.page.url)}&room_id=${encodeURIComponent(roomID)}`);
      webSocket.current.onmessage = event => handleMessage(event, widgets, setWidgets, terminalWidgets);
    }
  }, [page, widgets, props.page.url, roomID, terminalWidgets]);
  
  useEffect(() => {
    if (widgets) {
      terminalWidgets.current = terminalWidgets.current.slice(0, widgets.length);
    }
  }, [widgets]);
  
  useEffect(() => (() => {
    if (webSocket.current !== null) {
      console.log('Disconnecting from WebSocket ...');
      webSocket.current.close();
    }
  }), []);
  
  useEffect(() => {
    const fetchPage = async () => {
      const response = await fetch(`/api/page?page_url=${encodeURIComponent(props.page.url)}`);
      const respondedPage = await response.json();
      console.log(respondedPage);
      setPage(respondedPage);
      setWidgets(respondedPage.widgets.map(widget => {
        switch (widget.type) {
          case 'markdown':
            return widget;
          case 'text':
            return { ...widget, contents: null, error: null };
          case 'image':
            return { ...widget, contents: null, error: null };
          case 'button':
            return { ...widget, outputs: [] };
          case 'editor':
            return { ...widget, contents: null, error: null };
          case 'terminal':
            return widget;
          default:
            return widget;
        }
      }));
    };
    
    fetchPage();
  }, [props.page.url]);

  return (
    <div>
      <div className="centered title">{page.title}</div>
      <div className="centered">
        <div className="navigation">
          {props.parentPage !== null &&
            <div className="topics">
              <div className="label">Parent topic</div>
              <Link className="link up" to={props.parentPage.url}>
                {props.parentPage.title}
              </Link>
            </div>
          }
          {props.childrenPages !== null && props.childrenPages.length > 0 &&
            <div className="topics">
              <div className="label">Subtopics</div>
              {props.childrenPages.map(childPage =>
                <Link key={childPage.url} className="link down" to={childPage.url}>
                  {childPage.title}
                </Link>
              )}
            </div>
          }
        </div>
      </div>
      {widgets !== null && widgets.map((widget, i) => {
        switch (widget.type) {
          case 'markdown':
            return (<MarkdownWidget key={`${page.url}/${i}`} widget={widget} />);
          case 'text':
            return (<TextWidget key={`${page.url}/${i}`} widget={widget} />);
          case 'image':
            return (<ImageWidget key={`${page.url}/${i}`} widget={widget} />);
          case 'button':
            return (<ButtonWidget key={`${page.url}/${i}`} widget={widget} />);
          case 'editor':
            return (<EditorWidget key={`${page.url}/${i}`} widget={widget} />);
          case 'terminal':
            return (<TerminalWidget key={`${page.url}/${i}`} ref={element => terminalWidgets.current[i] = element} widget={widget} />);
          default:
            return (<div key={`${page.url}/${i}`} className="centered">Widget "{widget.type}" not implemented.</div>);
        }
      })}
    </div>
  );
}
