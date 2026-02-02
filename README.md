<div align="center">
    <img src="./docs/icon.png" alt="danlovesto.run logo" width="200" style="border-radius: 15px;" />
</div>

# danlovesto.run 🏃
Browser-only Strava dashboard that runs entirely in your local browser storage.

## What changed
This project has been rebuilt as a static, client-side app. The UI, dashboard rendering, and Strava API fetching happen locally in the browser. OAuth token exchanges are performed by a tiny server endpoint so you never expose your client secret in the UI.

## How to run
You can serve the files with any static server or open `index.html` directly.

```bash
# Example with python
python3 -m http.server 8080
```

Then open `http://localhost:8080` in your browser.

## Configure Strava
1. Create a Strava API application at https://www.strava.com/settings/api.
2. Add the **Authorization Callback Domain** for the domain you're hosting this from (localhost for local testing).
3. Use the redirect URL shown inside the app (it matches your current page URL).
4. Set the `CLIENT_ID` and `TOKEN_EXCHANGE_URL` constants in `app.js`.

## Token exchange service
Strava's OAuth token endpoint requires your client secret, so a server-side endpoint is required to exchange the authorization code for tokens. Keep the server minimal—just forward the `client_id`, `client_secret`, and OAuth payload to `https://www.strava.com/oauth/token` and return the response. Then point `TOKEN_EXCHANGE_URL` at that endpoint.

## References
- [Strava API Documentation](https://developers.strava.com/docs/reference/)
- [Strava API Playground](https://developers.strava.com/playground/)
