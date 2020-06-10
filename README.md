# interactive-markdown

## Development

The development environment consists of three components:

1. Backend server (for the dockerized backend)
2. Frontend server (for the React frontend with hot reloading)
3. NGINX reverse proxy (for proxying backend and frontend behind single HTTP server)

### Requirements

The backend server requires [Go](https://golang.org/) (written in version `1.14`, `>1.11` should work but not tested). Required modules are installed when running the backend server for the first time.

The frontend server requires [Node.JS](https://nodejs.org/) (written in `v14`, as of 2020-06-03 Node.JS `v8.10` and npm `v5.6` are required). It is required to install the required modules with `npm i` inside the `frontend/`-directory.

### Start

You may want to open three terminals, in each run one of the following commands:

#### Backend server

```bash
# Backend server
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
docker run --rm --net=host --name containerized-playground-nginx --mount type=bind,source="$(pwd)/docker/nginx-dev/nginx.conf",target=/etc/nginx/nginx.conf,readonly nginx
```

### Usage

Once started, head to `http://localhost/` to view the web application. Changes made in the frontend code recompile and reload the frontend page. Changes in the backend code require you to restart the backend server manually.
