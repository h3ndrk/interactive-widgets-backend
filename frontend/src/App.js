import React, { useEffect, useState } from 'react';
import { Route, Switch, useLocation, Redirect } from 'react-router-dom';
import Page from './page/Page';
import path from 'path-browserify';
import { v4 as uuidv4 } from 'uuid';

export default function App() {
  const [pages, setPages] = useState(null);
  const location = useLocation();
  useEffect(() => {
    const fetchPages = async () => {
      const response = await fetch('/api/pages');
      const respondedPages = await response.json();
      console.log(respondedPages);
      setPages(respondedPages);
    };
    fetchPages();
  }, []);
  
  // https://gist.github.com/johnelliott/cf77003f72f889abbc3f32785fa3df8d
  const hasUUIDHash = location.hash.match(/^#[0-9A-F]{8}-[0-9A-F]{4}-4[0-9A-F]{3}-[89AB][0-9A-F]{3}-[0-9A-F]{12}$/i);

  // sort URLs by length s.t. the longest match is rendered
  const pageRoutes = pages === null ? null : [...pages].sort((a, b) => b.url.length - a.url.length).map(page => {
    const parentURL = path.dirname(page.url) === page.url ? null : path.dirname(page.url);
    const parentPage = parentURL ? pages.find(page => page.url === parentURL) : null;
    const childrenPages = pages.filter(childPage => path.dirname(childPage.url) !== childPage.url && path.dirname(childPage.url) === page.url);

    return (<Route key={page.url} exact path={page.url}>
      {hasUUIDHash ? <Page page={page} parentPage={parentPage} childrenPages={childrenPages} /> : <Redirect to={`${page.url}#${uuidv4()}`} />}
    </Route>);
  });

  if (pageRoutes === null) {
    return (<div className="centered">Loading ...</div>);
  }

  return (
    <Switch>
      {pageRoutes}
    </Switch>
  );
}
