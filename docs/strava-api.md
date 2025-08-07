# Endpoints for Strava API
This document outlines the various endpoints for the Strava API.

## getAuthCode
> Open oauth website to fetch code

This opens the oauth link in the users default browser (linux only). Once the user logs in you need to grab the string `code=XXXX` from the url. Ignore any unable to connect messages.

```sh
    set -o allexport
    source ./.env
    set +o allexport

    AUTH_URL="https://www.strava.com/oauth/authorize?client_id=$CLIENT_ID&response_type=code&redirect_uri=http://localhost/exchange_token&approval_prompt=force&scope=read_all,profile:read_all,activity:read_all"
    echo "🔗 Opening: $AUTH_URL"
    nohup xdg-open "$AUTH_URL" >/dev/null 2>&1 &

    echo ""
    read -p "Paste code from URL here: " CODE

    # Remove any existing ACCESS_CODE entries
    sed -i '/^ACCESS_CODE=/d' .env
    echo "ACCESS_CODE=$CODE" >> .env
    echo "✅ ACCESS_CODE saved to .env"
```

## getAuthToken
> Authorizes the current user for the client id

Once the user has the oauth code, they need to fetch a valid token to be used in further requests. By default this will append the `access_token` and the `refresh_token` to the `.env`. The previous `ACCESS_CODE` is loaded from the `.env`.

**OPTIONS**
* raw
    * flags: -r --raw
    * desc: Prints the raw JSON result only

```sh
    set -o allexport
    source ./.env
    set +o allexport

    if [ -z "$ACCESS_CODE" ]; then
        echo "ACCESS_CODE not set in .env"
        exit 1
    fi

    response=$(curl -X POST https://www.strava.com/oauth/token \
      -d client_id=$CLIENT_ID \
      -d client_secret=$CLIENT_SECRET \
      -d code=$ACCESS_CODE \
      -d grant_type=authorization_code)

    if [ "$raw" = true ]; then
        echo "$response"
        exit 0
    fi

    # Parse and extract tokens
    access_token=$(echo "$response" | jq -r '.access_token')
    refresh_token=$(echo "$response" | jq -r '.refresh_token')

    if [ "$access_token" = "null" ] || [ "$refresh_token" = "null" ]; then
      echo "❌ Failed to retrieve tokens:"
      echo "$response"
      exit 1
    fi

    # Remove any existing ACCESS_TOKEN/REFRESH_TOKEN entries
    sed -i '/^ACCESS_TOKEN=/d' .env
    sed -i '/^REFRESH_TOKEN=/d' .env

    # Save new tokens to .env
    echo "ACCESS_TOKEN=$access_token" >> .env
    echo "REFRESH_TOKEN=$refresh_token" >> .env

    echo "$access_token"
```

## refreshAuthToken
> Refreshes the current users auth token

Used to refresh the `ACCESS_TOKEN` once it has expired rather than fetching a new token using the `ACCESS_CODE` from the oauth process.

**OPTIONS**
* raw
    * flags: -r --raw
    * desc: Prints the raw JSON result only

```sh
    set -o allexport
    source ./.env
    set +o allexport

    if [ -z "$REFRESH_TOKEN" ]; then
        echo "REFRESH_TOKEN not set in .env"
        exit 1
    fi

    response=$(curl -X POST https://www.strava.com/oauth/token \
      -d client_id=$CLIENT_ID \
      -d client_secret=$CLIENT_SECRET \
      -d code=$REFRESH_TOKEN \
      -d grant_type=refresh_token)

    if [ "$raw" = true ]; then
        echo "$response"
        exit 0
    fi

    # Parse and extract tokens
    access_token=$(echo "$response" | jq -r '.access_token')
    refresh_token=$(echo "$response" | jq -r '.refresh_token')

    if [ "$access_token" = "null" ] || [ "$refresh_token" = "null" ]; then
      echo "❌ Failed to retrieve tokens:"
      echo "$response"
      exit 1
    fi

    # Remove any existing ACCESS_TOKEN/REFRESH_TOKEN entries
    sed -i '/^ACCESS_TOKEN=/d' .env
    sed -i '/^REFRESH_TOKEN=/d' .env

    # Save new tokens to .env
    echo "ACCESS_TOKEN=$access_token" >> .env
    echo "REFRESH_TOKEN=$refresh_token" >> .env

    echo "$access_token"
```

## getAthlete
> Get athlete details for the currently logged in user

This endpoint fetches the athlete profile details for the user curently logged in via the `ACCESS_TOKEN`.

```sh
    set -o allexport
    source ./.env
    set +o allexport

    curl -X GET "https://www.strava.com/api/v3/athlete" \
      -H "Authorization: Bearer $ACCESS_TOKEN"
```

## getActivities

> Returns the activities of an athlete for a specific identifier. Requires activity:read.

Running a `GET /athlete/activities` fetches a list of the current users activities.

**OPTIONS**
* before
    * flags: -b --before
    * type: number
    * desc: An epoch timestamp to use for filtering activities that have taken place before a certain time.
* after
    * flags: -a --after
    * type: number
    * desc: An epoch timestamp to use for filtering activities that have taken place after a certain time.
* page
    * flags: -p --page
    * type: number
    * desc: Page number. Defaults to 1.
* per_page
    * flags: --per_page
    * type: number
    * desc: Number of items per page. Defaults to 30.

```sh
    set -o allexport
    source ./.env
    set +o allexport

    query=""
    [ -n "$before" ] && query+="&before=$before"
    [ -n "$after" ] && query+="&after=$after"
    [ -n "$page" ] && query+="&page=$page"
    [ -n "$per_page" ] && query+="&per_page=$per_page"

    if [ -n "$query" ]; then
      query="?${query#&}"
    fi

    url="https://www.strava.com/api/v3/athlete/activities$query"

    curl -X GET "$url" \
      -H "Authorization: Bearer $ACCESS_TOKEN"
```

## getActivityById (activityId)
> Get a DetailedActivity for a given activity id.

```sh
    set -o allexport
    source ./.env
    set +o allexport

    curl -X GET "https://www.strava.com/api/v3/athlete/$id" \
      -H "Authorization: Bearer $ACCESS_TOKEN"
```

