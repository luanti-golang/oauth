OAuth provider library for the luanti community

![](https://github.com/luanti-golang/oauth/workflows/test/badge.svg)

Docs: https://pkg.go.dev/github.com/luanti-golang/oauth

Example implementation: https://pkg.go.dev/github.com/luanti-golang/oauth#example-OauthHandler

# Overview

Supported providers:
* ContentDB
* Mesehub
* Github
* Discord

# Development

Create a "docker-compose.override.yml" with the following content:
```yml
version: "3.6"

services:
 oauth:
  environment:
   - GITHUB_APP_ID=
   - GITHUB_APP_SECRET=
   - MESEHUB_APP_ID=
   - MESEHUB_APP_SECRET=
   - DISCORD_APP_ID=
   - DISCORD_APP_SECRET=
   - CDB_APP_ID=
   - CDB_APP_SECRET=
```

Fill in the client-id's and secrets accordingly.

Callback-urls have to match the following list:
* ContentDB: `http://localhost:8080/api/oauth_callback/cdb`
* Mesehub: `http://localhost:8080/api/oauth_callback/mesehub`
* Github: `http://localhost:8080/api/oauth_callback/github`
* Discord: `http://localhost:8080/api/oauth_callback/discord`
* Codeberg: `http://localhost:8080/api/oauth_callback/codeberg`

Start the dev server:
```sh
docker-compose up
```

Navigate to http://localhost:8080 and try out the login links

# License

Code: `MIT`
