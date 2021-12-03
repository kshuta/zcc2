# Zendesk Coding Challenge 2022
This is Shuta Kumada's subsmission for the Zendesk Coding Challenge.

## Necessary Environment
- Go 1.17
- Modern web browser

## Installation
### 1. Clone the repository
```
git clone git@github.com:kshuta/zcc2.git
```

### 2. Make a `.env` file
Make a `.env` file within the root of the project directory and add the following environment variables:
- `API_TOKEN`: The api token to access your zendesk account. Make sure API access is enabled for your workspace.
- `API_EMAIL`: The email address you use to access your zendesk account.
- `API_SUBDOMAIN`: The subdomain for your zendesk workplace.
```
API_TOKEN=xxxxhdhdhdhxxxx
API_EMAIL=example@zendesk.com
API_SUBDOMAIN=zccchallenge
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
#### server.go
Responsible for serving the web pages with the ticket data.
#### tickets.go
Responsible for fetching ticket information from the zendesk api.

*please read the comments within the code for detailed documentation*

### Design documentation
#### Tests
##### Using regular expressions for assertion
To validate what is displayed on the browser, I have decided to use regular expressions for two reasons
1. It gives us flexibility in the html code we write, meaning we don't have to modify our tests everytime we give a slight change to the html file. (compared to using exact template matching)
2. For a project this size, the effort of validity was just right. Just checking for certain words in the html file is prone to false positives (e.g. we might be looking for a tag called "table", but we might also have an html element \<table\>. Thus, the test will say there is the word "table" in the html file even if the tag is not displayed). But writing a browser based test is a little over kill for a project this size.

Regular expressions gives us just the right balance between flexibility and validity.

##### tickets_test.go
The tickets.go file is not tested for two reasons (or a reason and an excuse).
1. The main functions of ticket.go relies on the zendesk api. Since zendesk api is already tested by other developers, there isn't a need to test them on my own.
2. I wasn't sure how I could have tested the other suppleental functions. Because the urls for fetching the api is generated using dynamic values from the environment variables, I wasn't sure how I would imitiate that in a test. As for the json parserers, I could have made a mock that imitates the functionality of the api, but I didn't have time to implement them (it's close to the finals in my college!).

#### Server
Instead of using a pre-built multiplexer, I decided to use regular expressions for routing, as there were only three paths that I had to consider.

The app will display errors from programs directly. I could not come up with an elegant way of creating user friendly error messages to be displayed on the screen, except for when the API returns a status code of something higher than 400.

#### Tickets
`TicketList` and `Tickets` are the two types responsible for storing ticket information. Most of the fields are derived from the fields included in the json format when fetching information from the api, but there are some custom fields to make front end navigation easier.







