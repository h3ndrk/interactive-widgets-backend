import React from 'react';

export default function ImageWidget(props) {
  return (
    <div className="centered image">
      <div className="file">{props.widget.file}</div>
      {props.widget.contents ?
        <img className="contents" src={`data:${props.widget.mime};base64,${props.widget.contents}`} alt={`Contents of ${props.widget.file}`} />
      :
        <div className="error">Error: {props.widget.error}</div>
      }
    </div>
  );
}
