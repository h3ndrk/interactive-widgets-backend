import React from 'react';
import MarkdownWidget from './widget/MarkdownWidget';

export default function StaticPage(props) {
  return (
    <div>
      {props.widgets !== null && props.widgets.map((widget, i) => {
        switch (widget.type) {
          case 'markdown':
            return (
              <MarkdownWidget
                key={`${props.page.url}/${i}`}
                widget={widget}
                pageURL={props.page.url}
              />
            );
          default:
            return (
              <div
                key={`${props.page.url}/${i}`}
                className="centered"
              >
                Widget "{widget.type}" not implemented.
              </div>
            );
        }
      })}
    </div>
  );
}
