import React from 'react';
import ReactMarkdown from 'react-markdown';

export default function MarkdownWidget(props) {
  return (
    <div className="centered markdown">
      <ReactMarkdown source={props.widget.contents} />
    </div>
  );
}
