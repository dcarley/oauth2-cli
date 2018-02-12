[![Build Status](https://travis-ci.org/dcarley/oauth2-cli.svg?branch=master)](https://travis-ci.org/dcarley/oauth2-cli)

# oauth2-cli

This is a small command line utility to get an OAuth access token for
three-legged flows where you authorize an application to access your
account, such as [Strava][].

[Strava]: http://strava.github.io/api/partner/v3/oauth/

It is useful for other command line utilities where you need an access token
but don't want to host the application on the web.

## Usage

Install:

    go get -u github.com/dcarley/oauth2-cli

Create an API application in the service of your choosing and set the
callback URL to as follows:

    http://localhost:8080/oauth/callback

Run with all of the necessary arguments, for example:

    $ oauth2-cli \
      -id REDACTED \
      -secret REDACTED \
      -auth https://www.strava.com/oauth/authorize \
      -token https://www.strava.com/oauth/token \
      -scope view_private

    Visit this URL in your browser:

    https://www.strava.com/oauth/authorize?access_type=offline&client_id=REDACTED&redirect_uri=http%3A%2F%2F127.0.0.1%3A8080%2Foauth%2Fcallback&response_type=code&scope=view_private&state=REDACTED

    ^C when finished.

Then follow the instructions in the CLI and subsequently your browser.

## Scopes

Multiple scopes can be given by specifying the argument multiple times:

    -scope read \
    -scope write \

Some services are lenient with their interpretation of the OAuth
specification so you will need to specify multiple scopes as a single comma
separated argument:

    -scope write,view_private
