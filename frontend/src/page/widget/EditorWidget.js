import React from 'react';
import AceEditor from 'react-ace';

export default function EditorWidget(props) {
  return (
    <div className="centered editor">
      <div className="file">{props.widget.file}</div>
      {props.widget.contents ?
        <AceEditor
          className="editor"
          name="ace-editor"
          mode=""
          theme=""
          value={props.widget.contents}
          onChange={props.onChange}
        />
      :
        <div className="error">Error: {props.widget.error}</div>
      }
    </div>
  );
}
