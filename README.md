# Zendesk Coding Challenge 2022
This is Shuta Kumada's subsmission for the Zendesk Coding Challenge.

## Necessary Environment
- Go 1.17

## Installation
### 1. Clone the repository
```
git clone git@github.com:kshuta/zcc2.git
```

### 2. Make a `.env` file
Make a `.env` file within the root of the project directory and add the following environmental variables:
- `API_TOKEN`: The api token to access your zendesk account. Make sure API access is enabled for your workspace.
- `API_EMAIL`: The email access you use to access your zendesk account.
- `API_SUBDOMAIN`: The subdomain for your zendesk workplace.
```
API_TOKEN=xxxxhdhdhdhxxxx
API_SUBDOMAIN=zccchallenge
API_EMAIL=example@zendesk.com
```

### 3. Run go build and start the application
```sh
go build .
./zcc2
# starts server on localhost:5000
```
Please do a `go install` of packages if prompted.

Default port will be `:5000`, but you can change this with the `-port` flag
```sh
./zcc2 -port :8888
# starts server on localhost:8888
```

## Application interfaces (routing)
| path            | function                                          |
| --------------- | ------------------------------------------------- |
| /               | Displays the ticket list                          |
| /tickets        | Displays the ticket list                          |
| /tickets/\[id\] | Displays ticket detail for ticket with \[id\] |

- Within the ticket list page, you can click on individual tickets to open the ticket detail page for that ticket. 

## Documentation
### Code documentation





