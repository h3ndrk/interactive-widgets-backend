import React from 'react';
import AceEditor from 'react-ace';
import seeNoEvilMonkeyEmoji from './see-no-evil-monkey.png';

export default function EditorWidget(props) {
  const showError = props.widget.error !== null || props.widget.contents === null;

  return (
    <div className="centered editor">
      <div className={showError ? 'contents with-error' : 'contents with-contents'}>
        {showError ?
          <div className="error">
            <img src={seeNoEvilMonkeyEmoji} alt="Oops" />
            <div className="title">Cannot view <code>{props.widget.file}</code></div>
            <div className="description">{props.widget.error !== null ? props.widget.error : 'There is no data'}</div>
          </div>
          :
          <AceEditor
            className="editor"
            name="ace-editor"
            mode=""
            theme=""
            width="100%"
            height="100%"
            value={props.widget.contents}
            onChange={props.onChange}
          />
        }
      </div>
      <div className="caption">Edit text contents of <code>{props.widget.file}</code></div>
    </div>
  );
}
