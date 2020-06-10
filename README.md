# inter-md

**Interactive Markdown pages backed by containers.**

*inter-md* (pronounced *interactive markdown*) provides a **web application server** that **serves easy-to-write [Markdown](https://daringfireball.net/projects/markdown/)** content to any modern webbrowser. In addition to the static Markdown, **interactive widgets** can be **embedded into the Markdown pages**.

With *inter-md*, the users can join rooms on the server in which widgets are executed, allowing connected users to interact with a page together. All widgets in one room are interconnected in the server backend and have a shared state. Since the widgets are executed in the server backend, the only required software for a user is an installed webbrowser.

The idea for *inter-md* came from the need of a simple software solution for building interactive learning websites. But because of its generic nature, it is not limited to that and can instead be used for many more use cases.

## Terminology

The **pages** are a hierarchical collection of multiple pages (i.e. a directory structure). A **page** consists of a Markdown document, additional assets. A **static page** has no widgets, an **interactive page** has at least one widget. When the user visits an interactive page it instantiates the page into a **room** (the last disconnecting user tears down a room).

## Widgets

The following widget types are supported and can be embedded into any Markdown page:

- **Button widget:** Clicking a button executes a pre-defined command in the server backend, which might change other widgets on the page.
- **Image widget:** Displays an live image from the page which also updates when another widget changes the image.
- **Text widget:** Displays live contents of a text file stored in a page which updates when the text file changes.
- **Editor widget:** Allows a user to change the contents of a text file.
- **Terminal widget:** Displays a remote terminal launched in the room of a page which also allows executing commands in the page.

## Pages

Currently *inter-md* uses a backend based on [docker](https://www.docker.com/) containers. If a page has at least one widget, it becomes an interactive page. Interactive pages define a `Dockerfile` which is used to instantiate docker containers for some widget types when the page is instantiated into a room. The `Dockerfile` allows the page's author to design and prepare an environment for the user to execute commands (e.g. pre-installing necessary tools, etc.). All executed widget containers have a shared volume mounted on `/data`.

You can write your own pages by creating a directory structure in the `pages/` directory of this repository. Page's authors can create any sub-directory structure (except a top-level `api` directory is reserved). The web application renders a navigation on each page to allow switching pages easily.

Each directory must have a `page.md` file containing the Markdown for the page and, if it is an interactive page, also a `Dockerfile`. Other assets (e.g. images) *can* be placed in the same directory. The `page.md` must begin with a level 1 heading (i.e. `# Title of the page`). You may want to keep this title short.

Widgets are embedded using HTML inside the Markdown content. Each widget must form its own paragraph (before and after a widget must be at least one free line). Only one widget is allowed in a paragraph. An example for a button widget surrounded by some Markdown content:

```md
# Here come widgets!

<x-button command="uname -a">Print server's architecture</x-button>

See above for some interactive button widget!
```

### Reference

In the following sections the widgets are documented with their available settings.

#### Button widget

The button widget shows a button with a given *label*. If clicked, the given *command* is executed in a new container. The container terminates once the command finishes. There can only be one running container at a time.

Syntax example:

```html
<x-button command="uname -a">Print server's architecture</x-button>
```

#### Image widget

The image widget shows an image based on the file contents of a given *file* path. It updates if the file contents changes. To correctly display the image in the webbrowser, the *[MIME](https://developer.mozilla.org/en-US/docs/Web/HTTP/Basics_of_HTTP/MIME_types)* type has to be given.

Syntax example:

```html
<x-image file="/data/image.png" mime="image/png" />
```

#### Text widget

The text widget shows an text based on the file contents of a given *file* path. It updates if the file contents changes.

Syntax example:

```html
<x-text file="/data/text.txt" />
```

#### Editor widget

The editor widget displays a text editor allowing editing the file contents of a given *file* path. It updates if the file contents changes.

Syntax example:

```html
<x-editor file="/data/text.txt" />
```

#### Terminal widget

The terminal widget shows a graphical pseudo terminal allowing executing any commands on a shell. It starts in a given *working-directory*. If the underlying terminal process terminates it is restarted.

Syntax example:

```html
<x-terminal working-directory="/data" />
```

## Deployment

The interactive Markdown pages are only tested on Linux currently and depend on the following software:

- [docker](https://www.docker.com/)
- [docker-compose](https://docs.docker.com/compose/)
- [make](https://www.gnu.org/software/make/)
- [wget](https://www.gnu.org/software/wget/)

Make sure to install them before proceeding.

Replace the contents of the `pages/` directory with your own interactive Markdown pages (maybe you want to create a separate git repository in there).

Create all mandatory docker images by executing `make docker-images`. After successfully building the images, the application can be started with `make up`. The interactive Markdown pages are now available at port `80` via any modern webbrowser.

If you make changes to the `pages/` directory's `Dockerfile`s and want to rebuild the images, run `make pages-images` (this is a sub-step of `make docker-images`).

Relevant files for the deployment are `Makefile` and all files in the `docker/` directory.

## Development

The development environment consists of three components:

1. Backend server (for the dockerized backend)
2. Frontend server (for the React frontend with hot reloading)
3. NGINX reverse proxy (for proxying backend and frontend behind single HTTP server)

### Requirements

The backend server requires [Go](https://golang.org/) (written in version `1.14`, `>1.11` should work but not tested). Required modules are installed when running the backend server for the first time.

The frontend server requires [Node.JS](https://nodejs.org/) (written in `v14`, as of 2020-06-03 Node.JS `v8.10` and npm `v5.6` are required). It is required to install the required modules with `npm i` inside the `frontend/`-directory.

### Preparation

The backend server needs a docker image `inter-md-monitor-write` to be built and read to run. In addition after each change of the page's `Dockerfile` the docker images for the pages need to be rebuilt.

```bash
# Build monitor-write docker image
make docker-monitor-write

# Build page's docker images
go run cmd/docker_build/main.go
```

### Start

You may want to open three terminals, in each run one of the following commands:

#### Backend server

```bash
# Backend server (required before: building docker containers, see above)
go run cmd/backend/main.go
```

#### Frontend server

```bash
# Frontend server (required before: `npm i`, see above)
cd frontend
BROWSER=none npm run start
```

#### NGINX reverse proxy

```bash
# NGINX reverse proxy
docker run --rm --net=host --name inter-md-nginx --mount type=bind,source="$(pwd)/docker/nginx-dev/nginx.conf",target=/etc/nginx/nginx.conf,readonly nginx
```

### Usage

Once started, head to `http://localhost/` to view the web application. Changes made in the frontend code recompile and reload the frontend page. Changes in the backend code require you to restart the backend server manually.

For debugging e.g. goroutine leaks, you can send a `SIGUSR1` signal to the backend process (when using `go run ...`, you need to send the signal to the `/tmp/go-build...` child process). This gives stack traces of all running goroutines.
