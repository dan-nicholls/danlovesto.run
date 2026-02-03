<div align="center">
    <img src="./docs/icon.png" alt="danlovesto.run logo" width="200" style="border-radius: 15px;" />
</div>

# danlovesto.run 🏃
Browser-only Strava dashboard that runs entirely in your local browser storage.

## What changed
This project has been rebuilt as a static, client-side app that runs with a tiny local server. The UI, dashboard rendering, and Strava API fetching happen locally in the browser. OAuth token exchanges are handled by the local server so you never expose your client secret in the UI.

## How to run
Create a local `.env` file based on the example and start the minimal server.

```bash
cp .env.example .env
node server.js
```

Then open `http://localhost:5173` in your browser.

## Configure Strava
1. Create a Strava API application at https://www.strava.com/settings/api.
2. Add the **Authorization Callback Domain** for the domain you're hosting this from (localhost for local testing).
3. Set the `STRAVA_CLIENT_ID` and `STRAVA_CLIENT_SECRET` values in your `.env`.

## Token exchange service
Strava's OAuth token endpoint requires your client secret, so a server-side endpoint is required to exchange the authorization code for tokens. The included `server.js` reads secrets from `.env`, serves the static UI, and proxies `/oauth/token` to Strava.

## References
- [Strava API Documentation](https://developers.strava.com/docs/reference/)
- [Strava API Playground](https://developers.strava.com/playground/)
