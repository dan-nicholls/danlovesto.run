const CLIENT_ID = "";
const TOKEN_EXCHANGE_URL = "";

const storageKeys = {
  accessToken: "strava_access_token",
  refreshToken: "strava_refresh_token",
  expiresAt: "strava_expires_at",
  oauthState: "strava_oauth_state",
};

const elements = {
  redirectUri: document.getElementById("redirectUri"),
  connectStrava: document.getElementById("connectStrava"),
  logout: document.getElementById("logout"),
  status: document.getElementById("status"),
  dashboard: document.getElementById("dashboard"),
  athleteName: document.getElementById("athleteName"),
  totalRuns: document.getElementById("totalRuns"),
  totalDistance: document.getElementById("totalDistance"),
  totalTime: document.getElementById("totalTime"),
  recentRuns: document.getElementById("recentRuns"),
};

const STRAVA_API = "https://www.strava.com/api/v3";
const STRAVA_OAUTH = "https://www.strava.com/oauth";

const redirectUri = `${window.location.origin}${window.location.pathname}`;
elements.redirectUri.value = redirectUri;

function setStatus(message) {
  elements.status.textContent = message;
}

function storeToken({ access_token, refresh_token, expires_at }) {
  localStorage.setItem(storageKeys.accessToken, access_token);
  localStorage.setItem(storageKeys.refreshToken, refresh_token);
  localStorage.setItem(storageKeys.expiresAt, String(expires_at));
}

function clearSession() {
  Object.values(storageKeys).forEach((key) => localStorage.removeItem(key));
  elements.dashboard.hidden = true;
  setStatus("Session cleared.");
}

function getToken() {
  const accessToken = localStorage.getItem(storageKeys.accessToken);
  const refreshToken = localStorage.getItem(storageKeys.refreshToken);
  const expiresAt = Number(localStorage.getItem(storageKeys.expiresAt));

  if (!accessToken || !refreshToken || !expiresAt) {
    return null;
  }

  return { accessToken, refreshToken, expiresAt };
}

function tokenExpired(token) {
  if (!token) return true;
  return Date.now() >= token.expiresAt * 1000 - 30000;
}

async function exchangeToken(payload) {
  if (!TOKEN_EXCHANGE_URL) {
    throw new Error(
      "Token exchange is not configured. Set TOKEN_EXCHANGE_URL in app.js to your server endpoint.",
    );
  }

  const response = await fetch(TOKEN_EXCHANGE_URL, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });

  if (!response.ok) {
    const errorText = await response.text();
    throw new Error(`Token exchange failed: ${errorText}`);
  }

  return response.json();
}

async function refreshAccessToken(token) {
  const refreshed = await exchangeToken({
    client_id: CLIENT_ID,
    grant_type: "refresh_token",
    refresh_token: token.refreshToken,
  });

  storeToken(refreshed);
  return {
    accessToken: refreshed.access_token,
    refreshToken: refreshed.refresh_token,
    expiresAt: refreshed.expires_at,
  };
}

async function fetchWithAuth(path) {
  let token = getToken();

  if (tokenExpired(token)) {
    token = await refreshAccessToken(token);
  }

  const response = await fetch(`${STRAVA_API}${path}`, {
    headers: {
      Authorization: `Bearer ${token.accessToken}`,
    },
  });

  if (!response.ok) {
    const errorText = await response.text();
    throw new Error(`Strava API error: ${errorText}`);
  }

  return response.json();
}

function formatDistance(meters) {
  if (!meters) return "0 km";
  const km = meters / 1000;
  return `${km.toFixed(1)} km`;
}

function formatDuration(seconds) {
  if (!seconds) return "0h";
  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  return `${hours}h ${minutes}m`;
}

function renderRecentRuns(activities) {
  elements.recentRuns.innerHTML = "";

  if (!activities.length) {
    elements.recentRuns.innerHTML = "<li>No recent runs found.</li>";
    return;
  }

  activities.forEach((activity) => {
    const item = document.createElement("li");
    const date = new Date(activity.start_date).toLocaleDateString();
    item.innerHTML = `
      <div>
        <strong>${activity.name}</strong>
        <span class="meta">${date}</span>
      </div>
      <div>
        <span>${formatDistance(activity.distance)}</span>
        <span class="meta">${formatDuration(activity.moving_time)}</span>
      </div>
    `;
    elements.recentRuns.appendChild(item);
  });
}

async function loadDashboard() {
  setStatus("Loading dashboard from Strava...");

  const athlete = await fetchWithAuth("/athlete");
  const stats = await fetchWithAuth(`/athletes/${athlete.id}/stats`);
  const activities = await fetchWithAuth("/athlete/activities?per_page=10");

  elements.athleteName.textContent = `${athlete.firstname} ${athlete.lastname}`;

  const runTotals = stats.all_run_totals || stats.recent_run_totals || {};
  elements.totalRuns.textContent = runTotals.count ?? "0";
  elements.totalDistance.textContent = formatDistance(runTotals.distance || 0);
  elements.totalTime.textContent = formatDuration(runTotals.moving_time || 0);

  renderRecentRuns(activities.filter((activity) => activity.type === "Run"));

  elements.dashboard.hidden = false;
  setStatus("Dashboard ready.");
}

function startAuthFlow() {
  if (!CLIENT_ID) {
    setStatus("Missing Strava client ID in app.js.");
    return;
  }

  const state = crypto.randomUUID();
  localStorage.setItem(storageKeys.oauthState, state);

  const params = new URLSearchParams({
    client_id: CLIENT_ID,
    redirect_uri: redirectUri,
    response_type: "code",
    scope: "read,activity:read_all",
    approval_prompt: "auto",
    state,
  });

  window.location.href = `${STRAVA_OAUTH}/authorize?${params.toString()}`;
}

async function handleAuthRedirect() {
  const urlParams = new URLSearchParams(window.location.search);
  const code = urlParams.get("code");
  const state = urlParams.get("state");
  const error = urlParams.get("error");

  if (error) {
    setStatus(`Authorization failed: ${error}`);
    return;
  }

  if (!code) return;

  const storedState = localStorage.getItem(storageKeys.oauthState);
  if (storedState && storedState !== state) {
    setStatus("State mismatch. Please retry authorization.");
    return;
  }

  setStatus("Exchanging code for access token...");

  try {
    const tokenResponse = await exchangeToken({
      client_id: CLIENT_ID,
      code,
      grant_type: "authorization_code",
    });

    storeToken(tokenResponse);
    window.history.replaceState({}, document.title, redirectUri);
    await loadDashboard();
  } catch (error) {
    setStatus(error.message);
  }
}

function setupEventListeners() {
  elements.connectStrava.addEventListener("click", startAuthFlow);
  elements.logout.addEventListener("click", clearSession);
}

async function init() {
  setupEventListeners();

  const token = getToken();
  if (token && !tokenExpired(token)) {
    await loadDashboard();
  } else {
    await handleAuthRedirect();
  }
}

init().catch((error) => {
  setStatus(`Error: ${error.message}`);
});
