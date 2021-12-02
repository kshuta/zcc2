# Zendesk Coding Challenge 2022
This is Shuta Kumada's subsmission for the Zendesk Coding Challenge.

## Necessary Environment
- Go 1.17

## Installation
1. Clone the repository
```
$ git clone git@github.com:kshuta/zcc2.git
```

2. Make a `.env` file
Make a `.env` file within the root of the project directory and add the following environmental variables:
- `API_TOKEN`: The api token to access your zendesk account. Make sure API access is enabled for your workspace.
- `API_EMAIL`: The email access you use to access your zendesk account.
- `API_SUBDOMAIN`: The subdomain for your zendesk workplace.
```
API_TOKEN=xxxxhdhdhdhxxxx
API_SUBDOMAIN=zccchallenge
API_EMAIL=example@zendesk.com
```

3. Run go build and start the application
```


