import React from 'react';

export default function EditorWidget(props) {
  return (
    <div className="centered editor">
      <div className="file">{props.widget.file}</div>
      {props.widget.contents ?
        <textarea className="contents">{props.widget.contents}</textarea>
      :
        <div className="error">Error: {props.widget.error}</div>
      }
    </div>
  );
}
