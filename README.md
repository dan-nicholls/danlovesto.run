<div align="center">
    <img src="./docs/icon.png" alt="danlovesto.run logo" width="200" style="border-radius: 15px;" />
</div>

# danlovesto.run 🏃
Browser-only Strava dashboard that runs entirely in your local browser storage.

## What changed
This project has been rebuilt as a static, client-side app. All Strava OAuth, token storage, and API fetching happens locally in the browser—there is no backend server at all.

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
4. Paste the Client ID and Client Secret into the form and connect.

## Important security note
Because this is fully browser-side, the Client Secret is stored in your browser storage so the app can exchange the OAuth code for a token. This keeps the project fully local and serverless, but it means you should only use this on a trusted device.

## References
- [Strava API Documentation](https://developers.strava.com/docs/reference/)
- [Strava API Playground](https://developers.strava.com/playground/)
