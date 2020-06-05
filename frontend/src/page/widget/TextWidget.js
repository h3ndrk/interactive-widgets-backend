import React from 'react';
import seeNoEvilMonkeyEmoji from './see-no-evil-monkey.png';

export default function TextWidget(props) {
  const showError = props.widget.error !== null || props.widget.contents === null;
  
  return (
    <div className="centered text">
      <div className={showError ? 'contents with-error' : 'contents with-contents'}>
        {showError ?
          <div className="error">
            <img src={seeNoEvilMonkeyEmoji} alt="Oops" />
            <div className="title">Cannot view <code>{props.widget.file}</code></div>
            <div className="description">{props.widget.error !== null ? props.widget.error : 'There is no data'}</div>
          </div>
        :
          props.widget.contents.split('\n').map((line, i) =>
            <React.Fragment key={i}>
              <div className="line-number">{i + 1}</div>
              <div className="line">{line}</div>
            </React.Fragment>
          )
        }
      </div>
      <div className="caption">Viewing text contents of <code>{props.widget.file}</code></div>
    </div>
  );
}
