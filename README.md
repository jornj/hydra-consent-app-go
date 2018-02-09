# hydra-consent-app-go

[![Build Status](https://travis-ci.org/ory/hydra-consent-app-go.svg?branch=master)](https://travis-ci.org/ory/hydra-consent-app-go)

This is a simple consent app for Hydra written in Go. It uses the Hydra SDK.
To run the example, first install Hydra, [dep](https://github.com/golang/dep)
and this project:

```sh
go get -u -d github.com/jornj/hydra-consent-app-go
cd $GOPATH/src/github.com/jornj/hydra-consent-app-go
dep ensure -v
```

Next open a shell and:

- run hydra from docker image (with memory only database; localhost port 4444)

```sh
# Download hydra image and run /bin/sh
docker run \
 -e "DATABASE_URL=memory" \
 -e "ISSUER=https://localhost:4444/" \
 -e "FORCE_ROOT_CLIENT_CREDENTIALS=demo:demo" \
 -e "CONSENT_URL=http://localhost:3000/consent" \
 -d --name my-hydra -p 4444:4444 \
 --entrypoint=/bin/sh -it \
 oryd/hydra

# Start the hydra host
hydra host --dangerous-force-http
```

In another console:

- Connect to the hydra server just started
- Create consent application and allow it to handle consents (localhost port 3000)
- Create client application (localhost port 3500)

```sh
# Get a shell in the Hydra docker image just started
docker exec -i -t my-hydra /bin/sh

# Connect to the Hydra
hydra connect --id demo --secret demo --url http://localhost:4444

# Create consent app
hydra clients create --skip-tls-verify \
  --id consent-app \
  --secret consent-secret \
  --name "Consent App Client" \
  --grant-types client_credentials \
  --response-types token \
  --allowed-scopes hydra.consent 

# Allow consent app to consent
hydra policies create --skip-tls-verify \ 
  --actions get,accept,reject \
  --description "Allow consent-app to manage OAuth2 consent requests." \
  --allow \
  --id consent-app-policy \
  --resources "rn:hydra:oauth2:consent:requests:<.*>" \
  --subjects consent-app

# Create client app with client specific callback url
hydra clients create --skip-tls-verify \
 --id client-app \
 --secret client-secret \
 -g authorization_code,refresh_token,client_credentials \
 -r token,code,id_token \
 --allowed-scopes openid,offline,demo \
 --callbacks http://localhost:3500/callback

```

In a second console, run the consent app

```sh
cd consent
go run consent.go
```

In a third console, run the client app

```sh
cd client
go run client.go
```

Then, open the browser:

```sh
open http://localhost:3500/
```

Now follow the steps described in the browser. If you encounter an error,
use the browser's back button to get back to the last screen.
