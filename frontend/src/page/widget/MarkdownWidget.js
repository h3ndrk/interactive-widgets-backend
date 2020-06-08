import React from 'react';
import ReactMarkdown from 'react-markdown';

export default function MarkdownWidget(props) {
  return (
    <div className="centered markdown">
      <ReactMarkdown
        source={props.widget.contents}
        transformImageUri={uri => `/api/page/image?page_url=${encodeURIComponent(props.pageURL)}&image_path=${encodeURIComponent(uri)}`}
      />
    </div>
  );
}
