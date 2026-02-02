const config = {
  clientId: "",
  tokenExchangeUrl: "",
};

const storageKeys = {
  accessToken: "strava_access_token",
  refreshToken: "strava_refresh_token",
  expiresAt: "strava_expires_at",
  oauthState: "strava_oauth_state",
};

const elements = {
  connectStrava: document.getElementById("connectStrava"),
  logout: document.getElementById("logout"),
  status: document.getElementById("status"),
  dashboard: document.getElementById("dashboard"),
  athleteName: document.getElementById("athleteName"),
  totalRuns: document.getElementById("totalRuns"),
  totalDistance: document.getElementById("totalDistance"),
  totalTime: document.getElementById("totalTime"),
  recentRuns: document.getElementById("recentRuns"),
  allRuns: document.getElementById("allRuns"),
  activityCount: document.getElementById("activityCount"),
};

const STRAVA_API = "https://www.strava.com/api/v3";
const STRAVA_OAUTH = "https://www.strava.com/oauth";

const redirectUri = `${window.location.origin}${window.location.pathname}`;

function setStatus(message) {
  elements.status.textContent = message;
}

function setAuthState(isAuthenticated) {
  elements.connectStrava.hidden = isAuthenticated;
  elements.logout.hidden = !isAuthenticated;
}

async function loadConfig() {
  try {
    const response = await fetch("/config");
    if (!response.ok) {
      throw new Error("Unable to load local configuration.");
    }
    const data = await response.json();
    config.clientId = data.clientId || "";
    config.tokenExchangeUrl = data.tokenExchangeUrl || "";
  } catch (error) {
    setStatus(`${error.message} Start the local server and try again.`);
  }
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
  setAuthState(false);
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
  if (!config.tokenExchangeUrl) {
    throw new Error(
      "Token exchange is not configured. Ensure the local server is running and /config is available.",
    );
  }

  const response = await fetch(config.tokenExchangeUrl, {
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
    client_id: config.clientId,
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

async function fetchAllActivities() {
  const allActivities = [];
  let page = 1;
  const perPage = 200;

  while (true) {
    const chunk = await fetchWithAuth(`/athlete/activities?per_page=${perPage}&page=${page}`);
    if (!chunk.length) break;
    allActivities.push(...chunk);
    if (chunk.length < perPage) break;
    page += 1;
  }

  return allActivities;
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

function hasPersonalRecord(activity) {
  const prCount = activity.pr_count ?? 0;
  const achievementCount = activity.achievement_count ?? 0;
  return prCount > 0 || achievementCount > 0;
}

function renderAllActivities(activities) {
  elements.allRuns.innerHTML = "";
  elements.activityCount.textContent = `${activities.length} activities loaded`;

  if (!activities.length) {
    elements.allRuns.innerHTML = "<li>No activities found.</li>";
    return;
  }

  activities.forEach((activity) => {
    const item = document.createElement("li");
    const date = new Date(activity.start_date).toLocaleDateString();
    const prLabel = hasPersonalRecord(activity) ? "<span class=\"badge\">PR</span>" : "";
    item.innerHTML = `
      <div>
        <strong>${activity.name}</strong>
        <span class="meta">${date}</span>
      </div>
      <div class="activity-meta">
        ${prLabel}
        <span>${formatDistance(activity.distance)}</span>
        <span class="meta">${formatDuration(activity.moving_time)}</span>
      </div>
    `;
    elements.allRuns.appendChild(item);
  });
}

async function loadDashboard() {
  setStatus("Loading dashboard from Strava...");

  const athlete = await fetchWithAuth("/athlete");
  const stats = await fetchWithAuth(`/athletes/${athlete.id}/stats`);
  const activities = await fetchAllActivities();

  elements.athleteName.textContent = `${athlete.firstname} ${athlete.lastname}`;

  const runTotals = stats.all_run_totals || stats.recent_run_totals || {};
  elements.totalRuns.textContent = runTotals.count ?? "0";
  elements.totalDistance.textContent = formatDistance(runTotals.distance || 0);
  elements.totalTime.textContent = formatDuration(runTotals.moving_time || 0);

  const runs = activities.filter((activity) => activity.type === "Run");
  renderRecentRuns(runs.slice(0, 10));
  renderAllActivities(activities);

  elements.dashboard.hidden = false;
  setStatus("Dashboard ready.");
  setAuthState(true);
}

function startAuthFlow() {
  if (!config.clientId) {
    setStatus("Missing Strava client ID. Check your local .env configuration.");
    return;
  }

  const state = crypto.randomUUID();
  localStorage.setItem(storageKeys.oauthState, state);

  const params = new URLSearchParams({
    client_id: config.clientId,
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
      client_id: config.clientId,
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
  await loadConfig();
  setupEventListeners();
  setAuthState(false);

  const token = getToken();
  if (token && !tokenExpired(token)) {
    try {
      await loadDashboard();
    } catch (error) {
      setStatus(`Error: ${error.message}`);
      setAuthState(false);
    }
  } else {
    await handleAuthRedirect();
  }
}

init().catch((error) => {
  setStatus(`Error: ${error.message}`);
});
