import React from 'react';
import seeNoEvilMonkeyEmoji from './see-no-evil-monkey.png';

export default function ImageWidget(props) {
  const showError = props.widget.error !== null || props.widget.contents === null;
  console.log('showError:', showError);
  const contentsStyle = showError ? {} : { backgroundImage: `url(data:${props.widget.mime};base64,${props.widget.contents})` };
  console.log('contentsStyle:', contentsStyle);

  return (
    <div className="centered image">
      <div className={showError ? 'contents with-error' : 'contents with-contents'} style={contentsStyle}>
        {showError &&
          <div className="error">
            <img src={seeNoEvilMonkeyEmoji} alt="Oops" />
            <div className="title">Cannot view <code>{props.widget.file}</code></div>
            <div className="description">{props.widget.error !== null ? props.widget.error : 'There is no data'}</div>
          </div>
        }
      </div>
      <div className="caption">Viewing image <code>{props.widget.file}</code></div>
    </div>
  );
}
