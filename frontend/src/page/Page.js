import React, { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import StaticPage from './StaticPage';
import InteractivePage from './InteractivePage';

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
          {props.parentPage !== null &&
            <div className="topics">
              <div className="label">Parent topic</div>
              <Link className="link up" to={props.parentPage.url}>
                {props.parentPage.title}
              </Link>
            </div>
          }
          {props.childrenPages !== null && props.childrenPages.length > 0 &&
            <div className="topics">
              <div className="label">Subtopics</div>
              {props.childrenPages.map(childPage =>
                <Link key={childPage.url} className="link down" to={childPage.url}>
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
