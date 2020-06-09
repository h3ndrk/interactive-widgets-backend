import React, { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import StaticPage from './StaticPage';
import InteractivePage from './InteractivePage';
import arrowBack from './arrow_back-black-18dp.svg';
import subdirectoryArrowRight from './subdirectory_arrow_right-black-18dp.svg';

export default function Page(props) {
  const [page, setPage] = useState(props.page);
  const [widgets, setWidgets] = useState(null);
  
  useEffect(() => {
    const fetchPage = async () => {
      const response = await fetch(`/api/page?page_url=${encodeURIComponent(props.page.url)}`);
      const respondedPage = await response.json();
      console.log(respondedPage);
      setPage(respondedPage);
      setWidgets(respondedPage.widgets.map(widget => {
        switch (widget.type) {
          case 'markdown':
            return widget;
          case 'text':
            return { ...widget, contents: null, error: null };
          case 'image':
            return { ...widget, contents: null, error: null };
          case 'button':
            return { ...widget, outputs: [] };
          case 'editor':
            return { ...widget, contents: null, error: null };
          case 'terminal':
            return widget;
          default:
            return widget;
        }
      }));
    };
    
    fetchPage();
  }, [props.page.url]);

  return (
    <>
      <div className="centered title">{page.title}</div>
      <div className="centered">
        <div className="navigation">
          {props.parentPages.length > 0 &&
            <div className="topics">
              <div className="label">Parent topic</div>
              <Link className="link" to={props.parentPages[props.parentPages.length - 1].url} style={{
                backgroundImage: `url(${arrowBack})`,
              }}>
                {props.parentPages[props.parentPages.length - 1].title}
              </Link>
            </div>
          }
          {props.childrenPages.length > 0 &&
            <div className="topics">
              <div className="label">Subtopics</div>
              {props.childrenPages.map(childPage =>
                <Link key={childPage.url} className="link" to={childPage.url} style={{
                  backgroundImage: `url(${subdirectoryArrowRight})`,
                }}>
                  {childPage.title}
                </Link>
              )}
            </div>
          }
        </div>
      </div>
      {widgets === null &&
        <div className="centered">Loading content ...</div>
      }
      {widgets !== null && !page.isInteractive &&
        <StaticPage page={page} widgets={widgets} />
      }
      {widgets !== null && page.isInteractive &&
        <InteractivePage page={page} widgets={widgets} />
      }
    </>
  );
}
