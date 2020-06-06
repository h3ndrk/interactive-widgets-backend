import React, { useState, useEffect } from 'react';
import AceEditor from 'react-ace';
import seeNoEvilMonkeyEmoji from './see-no-evil-monkey.png';

export default function EditorWidget(props) {
  const showError = props.widget.error !== null || props.widget.contents === null;
  const [contents, setContents] = useState("");

  useEffect(() => {
    if (props.widget.contents !== null) {
      setContents(props.widget.contents);
    }
  }, [props.widget.contents]);

  return (
    <div className="centered editor">
      <div className={showError ? 'contents with-error' : 'contents with-contents'}>
        <div className="buttons">
          <button onClick={() => {
            if (typeof props.onCreateOrEmpty === 'function') {
              props.onCreateOrEmpty();
            }
          }}>Create/Empty</button>
          {!showError &&
            <>
              <div className="spacer"></div>
              <button onClick={() => {
                if (typeof props.onChange === 'function') {
                  props.onChange(contents);
                }
              }}>Save</button>
              <div className="spacer"></div>
              <button onClick={() => {
                if (typeof props.onDelete === 'function') {
                  props.onDelete();
                }
              }}>Delete</button>
            </>
          }
        </div>
        {showError ?
          <div className="error-container">
            <div className="error">
              <img src={seeNoEvilMonkeyEmoji} alt="Oops" />
              <div className="title">Cannot view <code>{props.widget.file}</code></div>
              <div className="description">{props.widget.error !== null ? props.widget.error : 'There is no data'}</div>
            </div>
          </div>
          :
          <AceEditor
            className="editor"
            name="ace-editor"
            mode=""
            theme=""
            width="100%"
            height="100%"
            value={contents}
            onChange={value => {
              setContents(value);
            }}
          />
        }
      </div>
      <div className="caption">Edit text contents of <code>{props.widget.file}</code></div>
    </div>
  );
}
