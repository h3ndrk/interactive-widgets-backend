import React from 'react';

export default function TextWidget(props) {
  return (
    <div className="centered text">
      <div className="file">{props.widget.file}</div>
      {props.widget.contents ?
        <div className="contents">Contents: {props.widget.contents}</div>
      :
        <div className="error">Error: {props.widget.error}</div>
      }
    </div>
  );
}
