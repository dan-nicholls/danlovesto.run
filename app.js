const config = {
  clientId: "",
  tokenExchangeUrl: "",
  mapboxToken: "",
};

const storageKeys = {
  accessToken: "strava_access_token",
  refreshToken: "strava_refresh_token",
  expiresAt: "strava_expires_at",
  oauthState: "strava_oauth_state",
  cachedActivities: "strava_cached_activities",
  cachedActivitiesAt: "strava_cached_activities_at",
  cachedPrs: "strava_cached_prs",
  cachedPrsAt: "strava_cached_prs_at",
  cachedActivityDetails: "strava_cached_activity_details",
  cachedActivityDetailsAt: "strava_cached_activity_details_at",
};

const elements = {
  connectStrava: document.getElementById("connectStrava"),
  logout: document.getElementById("logout"),
  clearSession: document.getElementById("clearSession"),
  status: document.getElementById("status"),
  statusCard: document.getElementById("statusCard"),
  statusCountdown: document.getElementById("statusCountdown"),
  statusProgress: document.getElementById("statusProgress"),
  statusProgressTrack: document.getElementById("statusProgressTrack"),
  statusProgressBar: document.getElementById("statusProgressBar"),
  statusProgressLabel: document.getElementById("statusProgressLabel"),
  authCard: document.getElementById("authCard"),
  dashboard: document.getElementById("dashboard"),
  prTableCol1: document.getElementById("prTableCol1"),
  prTableCol2: document.getElementById("prTableCol2"),
  prTableCol3: document.getElementById("prTableCol3"),
  prMapTitle: document.getElementById("prMapTitle"),
  prDetailsTitle: document.getElementById("prDetailsTitle"),
  menuToggle: document.getElementById("menuToggle"),
  menuPanel: document.getElementById("menuPanel"),
  athleteName: document.getElementById("athleteName"),
  totalRuns: document.getElementById("totalRuns"),
  totalDistance: document.getElementById("totalDistance"),
  totalTime: document.getElementById("totalTime"),
  statsGrid: document.getElementById("statsGrid"),
  prTableBody: document.getElementById("prTableBody"),
  prMapContent: document.getElementById("prMapContent"),
  prDetailsContent: document.getElementById("prDetailsContent"),
  heatmaps: document.getElementById("heatmaps"),
  allRuns: document.getElementById("allRuns"),
  activityCount: document.getElementById("activityCount"),
};

const state = {
  prActivityIds: new Set(),
  isAuthenticated: false,
  progressTimer: null,
  prs: [],
  recentRuns: [],
  longestRuns: [],
  detailsMap: null,
  activeNotableTab: "prs",
};

const STRAVA_API = "https://www.strava.com/api/v3";
const STRAVA_OAUTH = "https://www.strava.com/oauth";

const redirectUri = `${window.location.origin}${window.location.pathname}`;

function setStatus(message) {
  elements.status.textContent = message;
  if (!elements.statusCard) return;
  const shouldHide =
    (message === "Dashboard ready." && !elements.dashboard.hidden) || !state.isAuthenticated;
  elements.statusCard.hidden = shouldHide;
}

function clearProgressTimer() {
  if (!state.progressTimer) return;
  clearInterval(state.progressTimer);
  state.progressTimer = null;
}

function hideStatusProgress() {
  clearProgressTimer();
  if (elements.statusProgress) {
    elements.statusProgress.hidden = true;
  }
}

function formatCountdown(seconds) {
  const minutes = Math.floor(seconds / 60);
  const remaining = seconds % 60;
  return `${minutes}:${String(remaining).padStart(2, "0")}`;
}

function setStatusProgress({ current, total, label, variant, indeterminate } = {}) {
  if (!elements.statusProgress || !elements.statusProgressBar || !elements.statusProgressTrack) return;
  if (variant !== "rate-limit") {
    clearProgressTimer();
    if (elements.statusCountdown) {
      elements.statusCountdown.textContent = "";
    }
  }
  elements.statusProgress.hidden = false;
  elements.statusProgressTrack.classList.toggle("is-indeterminate", Boolean(indeterminate));
  elements.statusProgressTrack.classList.toggle("is-rate-limit", variant === "rate-limit");

  if (typeof current === "number" && typeof total === "number" && total > 0) {
    const clamped = Math.min(current, total);
    const percent = Math.round((clamped / total) * 100);
    elements.statusProgressBar.style.width = `${percent}%`;
  } else if (!indeterminate) {
    elements.statusProgressBar.style.width = "0%";
  }

  if (elements.statusProgressLabel) {
    elements.statusProgressLabel.textContent = label || "";
  }
}

function startRateLimitCountdown(seconds) {
  clearProgressTimer();
  const totalSeconds = Math.max(1, Math.floor(seconds));
  let remaining = totalSeconds;
  setStatusProgress({
    current: remaining,
    total: totalSeconds,
    label: "Rate limits may take up to 15 minutes to reset.",
    variant: "rate-limit",
  });
  if (elements.statusCountdown) {
    elements.statusCountdown.textContent = formatCountdown(remaining);
  }

  state.progressTimer = setInterval(() => {
    remaining = Math.max(0, remaining - 1);
    setStatusProgress({
      current: remaining,
      total: totalSeconds,
      label: "Rate limits may take up to 15 minutes to reset.",
      variant: "rate-limit",
    });
    if (elements.statusCountdown) {
      elements.statusCountdown.textContent = formatCountdown(remaining);
    }
    if (remaining <= 0) {
      clearProgressTimer();
    }
  }, 1000);
}

function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

function setAuthState(isAuthenticated) {
  state.isAuthenticated = isAuthenticated;
  elements.connectStrava.hidden = isAuthenticated;
  if (elements.authCard) {
    elements.authCard.hidden = isAuthenticated;
  }
  if (elements.statusCard && !isAuthenticated) {
    elements.statusCard.hidden = true;
  }
  if (!isAuthenticated) {
    hideStatusProgress();
  } else {
    setStatus(elements.status.textContent || "");
  }
  if (elements.menuToggle) {
    elements.menuToggle.hidden = !isAuthenticated;
  }
  if (elements.menuPanel && !isAuthenticated) {
    elements.menuPanel.hidden = true;
  }
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
    config.mapboxToken = data.mapboxToken || "";
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
  state.prActivityIds.clear();
}

function getCachedItem(key, timestampKey, maxAgeMs) {
  const stored = localStorage.getItem(key);
  const storedAt = Number(localStorage.getItem(timestampKey));
  if (!stored || !storedAt) return null;
  try {
    const value = JSON.parse(stored);
    const isStale = Date.now() - storedAt > maxAgeMs;
    return { value, isStale };
  } catch {
    return null;
  }
}

function setCachedItem(key, timestampKey, value) {
  localStorage.setItem(key, JSON.stringify(value));
  localStorage.setItem(timestampKey, String(Date.now()));
}

function getCachedDetailsMap() {
  const cached = getCachedItem(
    storageKeys.cachedActivityDetails,
    storageKeys.cachedActivityDetailsAt,
    CACHE_MAX_AGE_MS,
  );
  if (!cached) return { map: {}, isStale: false };
  if (cached.isStale) {
    setStatus("Using cached activity details (stale). Clear session to refresh.");
  }
  return { map: cached.value || {}, isStale: cached.isStale };
}

function setCachedDetailsMap(map) {
  setCachedItem(storageKeys.cachedActivityDetails, storageKeys.cachedActivityDetailsAt, map);
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

async function fetchWithAuth(path, { retryCount = 0 } = {}) {
  let token = getToken();

  if (tokenExpired(token)) {
    token = await refreshAccessToken(token);
  }

  const response = await fetch(`${STRAVA_API}${path}`, {
    headers: {
      Authorization: `Bearer ${token.accessToken}`,
    },
  });

  if (response.status === 429 && retryCount < 3) {
    const retryAfter = Number(response.headers.get("retry-after"));
    const resetAt = Number(response.headers.get("x-rate-limit-reset"));
    let waitMs = 0;
    if (!Number.isNaN(retryAfter) && retryAfter > 0) {
      waitMs = retryAfter * 1000;
    } else if (!Number.isNaN(resetAt) && resetAt > 0) {
      waitMs = Math.max(resetAt * 1000 - Date.now(), 1000);
    } else {
      waitMs = 15 * 60 * 1000;
    }
    const countdownSeconds = Math.min(Math.ceil(waitMs / 1000), 90);
    setStatus("Rate limited. Waiting to retry...");
    startRateLimitCountdown(countdownSeconds);
    await sleep(countdownSeconds * 1000);
    return fetchWithAuth(path, { retryCount: retryCount + 1 });
  }

  if (response.status === 429) {
    const waitMs = 15 * 60 * 1000;
    const countdownSeconds = Math.min(Math.ceil(waitMs / 1000), 90);
    setStatus("Rate limited. Waiting to retry...");
    startRateLimitCountdown(countdownSeconds);
    await sleep(countdownSeconds * 1000);
    return fetchWithAuth(path, { retryCount: 0 });
  }

  if (!response.ok) {
    const errorText = await response.text();
    throw new Error(`Strava API error: ${errorText}`);
  }

  return response.json();
}

const CACHE_MAX_AGE_MS = 6 * 60 * 60 * 1000;

async function fetchAllActivities() {
  const cacheMaxAgeMs = CACHE_MAX_AGE_MS;
  const cached = getCachedItem(
    storageKeys.cachedActivities,
    storageKeys.cachedActivitiesAt,
    cacheMaxAgeMs,
  );
  if (cached && !cached.isStale) {
    return cached.value;
  }

  if (cached?.isStale) {
    setStatus("Refreshing activities from Strava...");
  }

  const allActivities = [];
  let page = 1;
  const perPage = 200;

  try {
    while (true) {
      const chunk = await fetchWithAuth(`/athlete/activities?per_page=${perPage}&page=${page}`);
      if (!chunk.length) break;
      allActivities.push(...chunk);
      if (chunk.length < perPage) break;
      page += 1;
    }

    setCachedItem(storageKeys.cachedActivities, storageKeys.cachedActivitiesAt, allActivities);
    return allActivities;
  } catch (error) {
    if (cached) {
      setStatus("Using cached activities (stale). Clear session to refresh.");
      return cached.value;
    }
    throw error;
  }
}

function formatDistance(meters) {
  if (!meters) return "0 km";
  const km = meters / 1000;
  return `${km.toFixed(1)} km`;
}

function formatDate(value) {
  if (!value) return "—";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "—";
  return date.toLocaleDateString(undefined, {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}

function formatDuration(seconds) {
  if (!seconds) return "0h";
  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  return `${hours}h ${minutes}m`;
}

function formatSeconds(seconds) {
  if (!seconds) return "0m";
  const hours = Math.floor(seconds / 3600);
  const minutes = Math.round((seconds % 3600) / 60);
  if (hours > 0) {
    return `${hours}h ${minutes}m`;
  }
  return `${minutes}m`;
}

function renderAxisLabels(labels = []) {
  if (!labels.length) return "";
  const maxLabels = 7;
  const step = labels.length > maxLabels ? Math.ceil(labels.length / maxLabels) : 1;
  const axis = labels
    .map((label, index) => {
      if (index % step !== 0) return "<span></span>";
      return `<span>${label}</span>`;
    })
    .join("");
  return `<div class="stats-chart-axis">${axis}</div>`;
}

function renderBarChart(values, labels, { showLabels = true } = {}) {
  if (!values.length) return "";
  const maxValue = Math.max(...values, 1);
  const barCount = values.length;
  const gap = 2;
  const barWidth = (100 - gap * (barCount - 1)) / barCount;
  const bars = values
    .map((value, index) => {
      const height = (value / maxValue) * 100;
      const x = index * (barWidth + gap);
      const label = labels?.[index] ?? "";
      const title = `${label}: ${value.toFixed(1)}`;
      return `
        <rect class="stats-bar-hit" x="${x}" y="0" width="${barWidth}" height="100">
          <title>${title}</title>
        </rect>
        <rect class="stats-bar" x="${x}" y="${100 - height}" width="${barWidth}" height="${height}" rx="2" />
      `;
    })
    .join("");

  return `
    <div class="stats-chart-wrap">
      <svg class="stats-chart" viewBox="0 0 100 100" preserveAspectRatio="none" aria-hidden="true">
        ${bars}
      </svg>
      ${showLabels ? renderAxisLabels(labels) : ""}
    </div>
  `;
}

function renderLineChart(points, { minLabel, maxLabel } = {}) {
  if (points.length < 2) return "";
  const maxY = Math.max(...points.map((p) => p.y), 1);
  const path = points
    .map((point, index) => {
      const x = (index / (points.length - 1)) * 100;
      const y = 100 - (point.y / maxY) * 100;
      return `${index === 0 ? "M" : "L"}${x.toFixed(2)},${y.toFixed(2)}`;
    })
    .join(" ");
  return `
    <div class="stats-chart-wrap">
      <svg class="stats-chart" viewBox="0 0 100 100" preserveAspectRatio="none" aria-hidden="true">
        <path class="stats-line" d="${path}" />
      </svg>
      <div class="stats-chart-axis">
        <span>${minLabel || ""}</span>
        <span>${maxLabel || ""}</span>
      </div>
    </div>
  `;
}

function renderDonutChart(values, labels) {
  const total = values.reduce((sum, value) => sum + value, 0) || 1;
  const radius = 18;
  const circumference = 2 * Math.PI * radius;
  let offset = 0;
  const palette = ["#8fb8ff", "#ff8b8b", "#f5c56b"];
  const arcs = values
    .map((value, index) => {
      const fraction = value / total;
      const length = fraction * circumference;
      const dash = `${length} ${circumference - length}`;
      const stroke = palette[index % palette.length];
      const label = labels[index] || "";
      const arc = `
        <circle
          class="stats-donut-arc"
          cx="24"
          cy="24"
          r="${radius}"
          stroke="${stroke}"
          stroke-dasharray="${dash}"
          stroke-dashoffset="${-offset}"
          data-label="${label}"
          data-value="${value}"
        />
      `;
      offset += length;
      return arc;
    })
    .join("");

  const legend = labels
    .map((label, index) => {
      return `
        <div class="stats-legend-item">
          <span class="stats-legend-dot" style="background:${palette[index % palette.length]}"></span>
          <span>${label}</span>
          <span class="stats-legend-value">${values[index] ?? 0}</span>
        </div>
      `;
    })
    .join("");

  return `
    <div class="stats-donut-wrap">
      <div class="stats-donut-shell">
        <svg class="stats-donut" viewBox="0 0 48 48" aria-hidden="true">
          <circle class="stats-donut-base" cx="24" cy="24" r="${radius}" />
          ${arcs}
        </svg>
        <div class="stats-donut-tooltip" hidden></div>
      </div>
      <div class="stats-legend">${legend}</div>
    </div>
  `;
}

function renderRadialTimeChart(values) {
  if (!values.length) return "";
  const smoothedValues = values.map((value, index) => {
    const prev = values[(index - 1 + values.length) % values.length] || 0;
    const next = values[(index + 1) % values.length] || 0;
    return (prev + value + next) / 3;
  });
  const maxValue = Math.max(...smoothedValues, 1);
  const total = values.reduce((sum, value) => sum + value, 0) || 1;
  const center = 50;
  const radius = 34;
  const levels = 4;
  const angleStep = (Math.PI * 2) / values.length;

  const grid = Array.from({ length: levels }, (_, level) => {
    const r = radius * ((level + 1) / levels);
    return `<circle class=\"stats-radial-grid\" cx=\"${center}\" cy=\"${center}\" r=\"${r.toFixed(2)}\" />`;
  }).join("");

  const gridLabels = Array.from({ length: levels }, (_, level) => {
    const r = radius * ((level + 1) / levels);
    const value = (maxValue * (level + 1)) / levels;
    const percent = Math.round((value / total) * 100);
    return `<text class=\"stats-radial-scale\" x=\"${center}\" y=\"${(center - r - 1).toFixed(2)}\">${percent}%</text>`;
  }).join("");

  const axisLabels = [
    { label: "12am", index: 0 },
    { label: "3am", index: 3 },
    { label: "6am", index: 6 },
    { label: "9am", index: 9 },
    { label: "12pm", index: 12 },
    { label: "3pm", index: 15 },
    { label: "6pm", index: 18 },
    { label: "9pm", index: 21 },
  ];

  const labelRadius = radius + 8;
  const labelNodes = axisLabels
    .map(({ label, index }) => {
      const angle = -Math.PI / 2 + index * angleStep;
      const x = center + labelRadius * Math.cos(angle);
      const y = center + labelRadius * Math.sin(angle);
      const anchor = Math.abs(Math.cos(angle)) < 0.2 ? "middle" : Math.cos(angle) > 0 ? "start" : "end";
      const dy = Math.sin(angle) > 0.4 ? "0.9em" : Math.sin(angle) < -0.4 ? "-0.2em" : "0.35em";
      return `<text class=\"stats-radial-label\" x=\"${x.toFixed(2)}\" y=\"${y.toFixed(2)}\" text-anchor=\"${anchor}\" dy=\"${dy}\">${label}</text>`;
    })
    .join("");

  const areaPoints = smoothedValues
    .map((value, index) => {
      const angle = -Math.PI / 2 + index * angleStep;
      const r = (value / maxValue) * radius;
      const x = center + r * Math.cos(angle);
      const y = center + r * Math.sin(angle);
      return { x, y };
    });

  const areaPath = buildSmoothClosedPath(areaPoints, 0.7);

  return `
    <div class="stats-chart-wrap stats-chart-wrap--radial">
      <svg class="stats-chart stats-chart--radial" viewBox="0 0 100 100" aria-hidden="true">
        ${grid}
        ${gridLabels}
        <path class="stats-radial-area" d="${areaPath} Z" />
        ${labelNodes}
      </svg>
    </div>
  `;
}

function buildSmoothClosedPath(points, tension = 0.6) {
  if (points.length < 2) return "";
  const count = points.length;
  const path = [];
  for (let i = 0; i < count; i += 1) {
    const p0 = points[(i - 1 + count) % count];
    const p1 = points[i];
    const p2 = points[(i + 1) % count];
    const p3 = points[(i + 2) % count];
    const cp1x = p1.x + ((p2.x - p0.x) * tension) / 6;
    const cp1y = p1.y + ((p2.y - p0.y) * tension) / 6;
    const cp2x = p2.x - ((p3.x - p1.x) * tension) / 6;
    const cp2y = p2.y - ((p3.y - p1.y) * tension) / 6;
    if (i === 0) {
      path.push(`M${p1.x.toFixed(2)},${p1.y.toFixed(2)}`);
    }
    path.push(
      `C${cp1x.toFixed(2)},${cp1y.toFixed(2)} ${cp2x.toFixed(2)},${cp2y.toFixed(2)} ${p2.x.toFixed(2)},${p2.y.toFixed(2)}`,
    );
  }
  return path.join(" ");
}

function renderRadarChart(values, labels) {
  if (!values.length) return "";
  const maxValue = Math.max(...values, 1);
  const count = values.length;
  const center = 50;
  const radius = 38;
  const levels = 4;
  const angleStep = (Math.PI * 2) / count;

  const grid = Array.from({ length: levels }, (_, level) => {
    const r = radius * ((level + 1) / levels);
    const points = values
      .map((_, index) => {
        const angle = -Math.PI / 2 + index * angleStep;
        const x = center + r * Math.cos(angle);
        const y = center + r * Math.sin(angle);
        return `${x.toFixed(2)},${y.toFixed(2)}`;
      })
      .join(" ");
    return `<polygon class=\"stats-radar-grid\" points=\"${points}\" />`;
  }).join("");

  const scaleLabels = Array.from({ length: levels }, (_, level) => {
    const r = radius * ((level + 1) / levels);
    const value = (maxValue * (level + 1)) / levels;
    return `<text class=\"stats-radar-scale\" x=\"${center}\" y=\"${(center - r + 2).toFixed(2)}\">${value.toFixed(1)} km</text>`;
  }).join("");

  const axes = values
    .map((_, index) => {
      const angle = -Math.PI / 2 + index * angleStep;
      const x = center + radius * Math.cos(angle);
      const y = center + radius * Math.sin(angle);
      return `<line class=\"stats-radar-axis\" x1=\"${center}\" y1=\"${center}\" x2=\"${x.toFixed(2)}\" y2=\"${y.toFixed(2)}\" />`;
    })
    .join("");

  const shapePoints = values
    .map((value, index) => {
      const angle = -Math.PI / 2 + index * angleStep;
      const r = (value / maxValue) * radius;
      const x = center + r * Math.cos(angle);
      const y = center + r * Math.sin(angle);
      return `${x.toFixed(2)},${y.toFixed(2)}`;
    })
    .join(" ");

  const labelRadius = radius + 8;
  const labelNodes = labels
    .map((label, index) => {
      const angle = -Math.PI / 2 + index * angleStep;
      const x = center + labelRadius * Math.cos(angle);
      const y = center + labelRadius * Math.sin(angle);
      const anchor = Math.abs(Math.cos(angle)) < 0.2 ? "middle" : Math.cos(angle) > 0 ? "start" : "end";
      const dy = Math.sin(angle) > 0.4 ? "0.9em" : Math.sin(angle) < -0.4 ? "-0.2em" : "0.35em";
      return `<text class=\"stats-radar-label\" x=\"${x.toFixed(2)}\" y=\"${y.toFixed(2)}\" text-anchor=\"${anchor}\" dy=\"${dy}\">${label}</text>`;
    })
    .join("");

  return `
    <div class="stats-chart-wrap stats-chart-wrap--radar">
      <svg class="stats-chart stats-chart--radar" viewBox="0 0 100 100" aria-hidden="true" data-values='${JSON.stringify(values)}' data-labels='${JSON.stringify(labels)}'>
        ${grid}
        ${axes}
        ${scaleLabels}
        <polygon class="stats-radar-shape" points="${shapePoints}" />
        ${labelNodes}
      </svg>
      <div class="stats-radar-tooltip" hidden></div>
    </div>
  `;
}

async function mapWithConcurrency(items, limit, mapper) {
  if (!items.length) return [];
  const results = new Array(items.length);
  let nextIndex = 0;
  let inFlight = 0;

  return new Promise((resolve, reject) => {
    const launch = () => {
      while (inFlight < limit && nextIndex < items.length) {
        const currentIndex = nextIndex;
        nextIndex += 1;
        inFlight += 1;
        Promise.resolve(mapper(items[currentIndex], currentIndex))
          .then((result) => {
            results[currentIndex] = result;
            inFlight -= 1;
            if (nextIndex >= items.length && inFlight === 0) {
              resolve(results);
              return;
            }
            launch();
          })
          .catch((error) => {
            reject(error);
          });
      }
    };

    launch();
  });
}

function countWeekdays(startDate, endDate) {
  const counts = Array.from({ length: 7 }, () => 0);
  if (!startDate || !endDate) return counts;
  const current = new Date(Date.UTC(startDate.getFullYear(), startDate.getMonth(), startDate.getDate()));
  const end = new Date(Date.UTC(endDate.getFullYear(), endDate.getMonth(), endDate.getDate()));
  while (current <= end) {
    const weekday = (current.getUTCDay() + 6) % 7;
    counts[weekday] += 1;
    current.setUTCDate(current.getUTCDate() + 1);
  }
  return counts;
}

function bucketize(value, thresholds) {
  for (let i = 0; i < thresholds.length; i += 1) {
    if (value <= thresholds[i]) return i;
  }
  return thresholds.length;
}

function renderStats(runs) {
  if (!elements.statsGrid) return;
  if (!runs.length) {
    elements.statsGrid.innerHTML = "<div class=\"stats-card\">No stats yet.</div>";
    return;
  }

  const totalDistance = runs.reduce((sum, run) => sum + (run.distance || 0), 0);
  const totalTime = runs.reduce((sum, run) => sum + (run.moving_time || 0), 0);
  const avgPace = totalDistance > 0 ? totalTime / (totalDistance / 1000) : 0;

  const distanceByYear = new Map();
  const distanceByWeekday = Array.from({ length: 7 }, () => 0);
  const runsByWeekday = Array.from({ length: 7 }, () => 0);
  const distanceByMonth = Array.from({ length: 12 }, () => 0);

  const distanceBins = [1, 3, 5, 6, 8, 10];
  const distanceBinCounts = Array.from({ length: distanceBins.length + 1 }, () => 0);

  const paceValues = [];

  const detailsCache = getCachedDetailsMap();
  const detailsMap = detailsCache.map;
  const friendCounts = { with: 0, solo: 0, unknown: 0 };
  const tempBins = [0, 5, 10, 15, 20, 25];
  const tempBinCounts = Array.from({ length: tempBins.length + 1 }, () => 0);
  let tempUnknown = 0;
  const weatherCounts = { sun: 0, cloud: 0, rain: 0, unknown: 0 };

  const timeOfDayCounts = { morning: 0, afternoon: 0, evening: 0, night: 0 };
  const elevationByYear = new Map();

  const getLocalHour = (run) => {
    const local = run.start_date_local || run.start_date;
    const match = typeof local === "string" ? local.match(/T(\d{2}):/) : null;
    if (match) return Number(match[1]);
    const date = new Date(local);
    return date.getHours();
  };

  runs.forEach((run) => {
    const date = new Date(run.start_date_local);
    const year = date.getFullYear();
    const distanceKm = (run.distance || 0) / 1000;
    const weekday = (date.getDay() + 6) % 7;
    const pace = run.distance ? (run.moving_time || 0) / distanceKm / 60 : 0;

    distanceByYear.set(year, (distanceByYear.get(year) || 0) + distanceKm);
    distanceByWeekday[weekday] += distanceKm;
    runsByWeekday[weekday] += 1;
    distanceByMonth[date.getMonth()] += distanceKm;

    distanceBinCounts[bucketize(distanceKm, distanceBins)] += 1;
    if (pace > 0) {
      paceValues.push(pace);
    }

    const hour = getLocalHour(run);
    if (hour >= 5 && hour < 12) timeOfDayCounts.morning += 1;
    else if (hour >= 12 && hour < 17) timeOfDayCounts.afternoon += 1;
    else if (hour >= 17 && hour < 21) timeOfDayCounts.evening += 1;
    else timeOfDayCounts.night += 1;

    elevationByYear.set(
      year,
      (elevationByYear.get(year) || 0) + (run.total_elevation_gain || 0),
    );

    const details = detailsMap[run.id];
    if (details?.athlete_count) {
      if (details.athlete_count > 1) friendCounts.with += 1;
      else friendCounts.solo += 1;
    } else {
      friendCounts.unknown += 1;
    }

    if (typeof details?.average_temp === "number") {
      tempBinCounts[bucketize(details.average_temp, tempBins)] += 1;
    } else {
      tempUnknown += 1;
    }

    const weather = details?.weather ?? details?.weather_type ?? "unknown";
    if (weather === "sun" || weather === "clear") weatherCounts.sun += 1;
    else if (weather === "cloud" || weather === "cloudy") weatherCounts.cloud += 1;
    else if (weather === "rain" || weather === "rainy") weatherCounts.rain += 1;
    else weatherCounts.unknown += 1;
  });

  const years = Array.from(distanceByYear.keys()).sort((a, b) => a - b);
  const yearValues = years.map((year) => distanceByYear.get(year));
  const yearLabels = years.map((year) => String(year));

  const weekdayLabels = ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"];
  const avgKmByWeekday = distanceByWeekday.map((distance, index) => {
    const runCount = runsByWeekday[index] || 1;
    return distance / runCount;
  });

  const distanceBinLabels = ["0-1", "1-3", "3-5", "5-6", "6-8", "8-10", "10+"].map(
    (label) => `${label} km`,
  );
  const paceMin = paceValues.length ? Math.min(...paceValues) : 0;
  const paceMax = paceValues.length ? Math.max(...paceValues) : 0;
  const paceSamples = 30;
  const paceRange = paceMax - paceMin || 1;
  const paceStep = paceRange / (paceSamples - 1);
  const bandwidth = Math.max(0.25, paceRange / 12);
  const paceDistribution = Array.from({ length: paceSamples }, (_, index) => {
    const x = paceMin + paceStep * index;
    const y = paceValues.reduce((sum, value) => {
      const z = (value - x) / bandwidth;
      return sum + Math.exp(-0.5 * z * z);
    }, 0);
    return { x, y };
  });
  const tempBinLabels = ["<0", "0-5", "5-10", "10-15", "15-20", "20-25", "25+"].map(
    (label) => `${label}°C`,
  );

  const filterUnknown = (labels, values) => {
    return labels.reduce(
      (acc, label, index) => {
        if (label === "Unknown" && (values[index] ?? 0) === 0) return acc;
        acc.labels.push(label);
        acc.values.push(values[index]);
        return acc;
      },
      { labels: [], values: [] },
    );
  };

  const friendData = filterUnknown(
    ["With friends", "Solo", "Unknown"],
    [friendCounts.with, friendCounts.solo, friendCounts.unknown],
  );

  const weatherData = filterUnknown(
    ["Sun", "Cloud", "Rain", "Unknown"],
    [weatherCounts.sun, weatherCounts.cloud, weatherCounts.rain, weatherCounts.unknown],
  );

  const tempData = filterUnknown(
    [...tempBinLabels, "Unknown"],
    [...tempBinCounts, tempUnknown],
  );

  const hourCounts = Array.from({ length: 24 }, () => 0);
  runs.forEach((run) => {
    const hour = getLocalHour(run);
    hourCounts[hour] += 1;
  });

  const elevValues = years.map((year) => (elevationByYear.get(year) || 0) / 1000);
  const stats = [
    {
      title: "Distance per year",
      value: `${years.length} yrs`,
      chart: renderBarChart(yearValues, yearLabels),
    },
    {
      title: "Avg km per day",
      value: `${(totalDistance / 1000 / runs.length).toFixed(1)} km`,
      chart: renderRadarChart(avgKmByWeekday, weekdayLabels),
    },
    {
      title: "Run distance bins",
      value: `${runs.length} runs`,
      chart: renderBarChart(distanceBinCounts, distanceBinLabels),
    },
    {
      title: "Pace distribution",
      value: avgPace ? `${avgPace.toFixed(1)} min/km` : "—",
      chart: renderLineChart(paceDistribution, {
        minLabel: `${paceMin.toFixed(1)} min/km`,
        maxLabel: `${paceMax.toFixed(1)} min/km`,
      }),
    },
    {
      title: "Runs with friends",
      value: `${friendCounts.with} group`,
      chart: renderDonutChart(friendData.values, friendData.labels),
    },
    {
      title: "Temp vs runs",
      value: tempUnknown ? "Partial" : "All runs",
      chart: renderBarChart(tempData.values, tempData.labels),
    },
    {
      title: "Weather vs runs",
      value: weatherCounts.unknown ? "Partial" : "All runs",
      chart: renderBarChart(weatherData.values, weatherData.labels),
    },
    {
      title: "Runs by time of day",
      value: `${runs.length} runs`,
      chart: renderRadialTimeChart(hourCounts),
    },
    {
      title: "Elevation per year",
      value: "km gain",
      chart: renderBarChart(elevValues, yearLabels),
    },
  ];

  elements.statsGrid.innerHTML = stats
    .map((stat) => `
      <div class="stats-card">
        <div class="stats-card-header">
          <div class="stats-title">${stat.title}</div>
          <div class="stats-subtitle">${stat.value}</div>
        </div>
        ${stat.chart || ""}
      </div>
    `)
    .join("");

  setupRadarTooltips();
  setupDonutTooltips();
}

function setupRadarTooltips() {
  document.querySelectorAll(".stats-chart-wrap--radar").forEach((wrap) => {
    const svg = wrap.querySelector(".stats-chart--radar");
    const tooltip = wrap.querySelector(".stats-radar-tooltip");
    if (!svg || !tooltip) return;

    const values = JSON.parse(svg.getAttribute("data-values") || "[]");
    const labels = JSON.parse(svg.getAttribute("data-labels") || "[]");
    if (!values.length) return;

    const center = 50;
    const radius = 38;
    const angleStep = (Math.PI * 2) / values.length;
    const points = values.map((value, index) => {
      const angle = -Math.PI / 2 + index * angleStep;
      const r = (value / Math.max(...values, 1)) * radius;
      return {
        x: center + r * Math.cos(angle),
        y: center + r * Math.sin(angle),
        label: labels[index] || "",
        value,
      };
    });

    const handleMove = (event) => {
      const rect = svg.getBoundingClientRect();
      const x = ((event.clientX - rect.left) / rect.width) * 100;
      const y = ((event.clientY - rect.top) / rect.height) * 100;
      let closest = null;
      let bestDistance = Infinity;
      points.forEach((point) => {
        const dx = point.x - x;
        const dy = point.y - y;
        const distance = dx * dx + dy * dy;
        if (distance < bestDistance) {
          bestDistance = distance;
          closest = point;
        }
      });
      if (!closest) return;

      tooltip.hidden = false;
      tooltip.textContent = `${closest.label}: ${closest.value.toFixed(1)} km`;
      const wrapRect = wrap.getBoundingClientRect();
      const left = ((closest.x / 100) * wrapRect.width);
      const top = ((closest.y / 100) * wrapRect.height);
      tooltip.style.left = `${left}px`;
      tooltip.style.top = `${top}px`;
    };

    const hide = () => {
      tooltip.hidden = true;
    };

    svg.addEventListener("mousemove", handleMove);
    svg.addEventListener("mouseleave", hide);
  });
}

function setupDonutTooltips() {
  document.querySelectorAll(".stats-donut-shell").forEach((shell) => {
    const tooltip = shell.querySelector(".stats-donut-tooltip");
    const arcs = shell.querySelectorAll(".stats-donut-arc");
    if (!tooltip || !arcs.length) return;

    const show = (event) => {
      const target = event.currentTarget;
      if (!(target instanceof SVGElement)) return;
      const label = target.getAttribute("data-label") || "";
      const value = target.getAttribute("data-value") || "0";
      tooltip.textContent = `${label}: ${value}`;
      tooltip.hidden = false;

      const rect = shell.getBoundingClientRect();
      const x = event.clientX - rect.left;
      const y = event.clientY - rect.top;
      tooltip.style.left = `${x}px`;
      tooltip.style.top = `${y}px`;
    };

    const hide = () => {
      tooltip.hidden = true;
    };

    arcs.forEach((arc) => {
      arc.addEventListener("mousemove", show);
      arc.addEventListener("mouseleave", hide);
    });
  });
}


function setNotableTab(tab) {
  state.activeNotableTab = tab;
  document.querySelectorAll(".notable-tabs .tab-btn").forEach((button) => {
    button.classList.toggle("is-active", button.dataset.tab === tab);
  });

  if (elements.prTableCol1) {
    elements.prTableCol1.textContent = tab === "prs" ? "Distance" : "Rank";
  }
  if (elements.prTableCol2) {
    elements.prTableCol2.textContent = "Date";
  }
  if (elements.prTableCol3) {
    elements.prTableCol3.textContent = tab === "prs" ? "Time" : "Distance";
  }
  if (elements.prMapTitle) {
    elements.prMapTitle.textContent = tab === "prs" ? "Route map" : "Route map & heart rate";
  }
  if (elements.prDetailsTitle) {
    elements.prDetailsTitle.textContent = tab === "prs" ? "Run details" : "Activity details";
  }

  if (tab === "recent") {
    renderRecentRunsTable(state.recentRuns);
  } else if (tab === "longest") {
    renderRecentRunsTable(state.longestRuns);
  } else {
    renderPRs(state.prs);
  }
}

async function getActivityDetails(activityId) {
  if (!state.detailsMap) {
    state.detailsMap = getCachedDetailsMap().map;
  }
  if (state.detailsMap[activityId]) return state.detailsMap[activityId];
  const details = await fetchWithAuth(`/activities/${activityId}`);
  state.detailsMap[activityId] = details;
  setCachedDetailsMap(state.detailsMap);
  return details;
}

async function renderRecentRunSelection(activity) {
  elements.prMapContent.textContent = "Loading route map...";
  elements.prDetailsContent.textContent = "Loading activity details...";

  const details = await getActivityDetails(activity.id);
  const heartRate = details.average_heartrate
    ? `${Math.round(details.average_heartrate)} bpm`
    : "—";
  const elevation = details.total_elevation_gain
    ? `${Math.round(details.total_elevation_gain)} m`
    : "—";

  elements.prMapContent.innerHTML = renderMapboxStaticMap(details.map?.summary_polyline || "");

  const date = formatDate(details.start_date);
  elements.prDetailsContent.innerHTML = `
    <div class="pr-details-meta">
      <span class="meta">${details.id}</span>
      <span class="meta">${date}</span>
    </div>
    <div class="pr-details-hero">
      <div class="pr-details-title">
        <a class="activity-link" href="https://www.strava.com/activities/${details.id}" target="_blank" rel="noopener noreferrer">
          ${details.name || activity.name || "Run"}
          <svg viewBox="0 0 24 24" aria-hidden="true">
            <path d="M14 5h5v5" />
            <path d="M10 14L19 5" />
            <path d="M19 14v5h-9a2 2 0 0 1-2-2v-9" />
          </svg>
        </a>
      </div>
    </div>
    <div class="pr-metrics">
      <div class="pr-metric">
        <span class="label">Distance</span>
        <span class="value">${formatDistance(details.distance)}</span>
      </div>
      <div class="pr-metric">
        <span class="label">Time</span>
        <span class="value">${formatTime(details.moving_time)}</span>
      </div>
      <div class="pr-metric">
        <span class="label">Pace</span>
        <span class="value">${formatPace(details.moving_time, details.distance)}</span>
      </div>
      <div class="pr-metric">
        <span class="label">Elevation</span>
        <span class="value">${elevation}</span>
      </div>
      <div class="pr-metric">
        <span class="label">Weather</span>
        <span class="value">—</span>
      </div>
      <div class="pr-metric">
        <span class="label">Location</span>
        <span class="value">—</span>
      </div>
    </div>
  `;
}

function renderRecentRunsTable(activities) {
  elements.prTableBody.innerHTML = "";

  if (!activities.length) {
    elements.prTableBody.innerHTML = "<div class=\"pr-table-empty\">No recent runs found.</div>";
    elements.prMapContent.textContent = "Select a run to view the map.";
    elements.prDetailsContent.textContent = "Select a run to view activity details.";
    return;
  }

  activities.forEach((activity, index) => {
    const row = document.createElement("button");
    row.type = "button";
    row.className = "pr-table-row";
    row.dataset.index = String(index);
    const date = formatDate(activity.start_date);
    row.innerHTML = `
      <span>#${index + 1}</span>
      <span class="meta">${date}</span>
      <span>${formatDistance(activity.distance)}</span>
    `;
    row.addEventListener("click", () => {
      document.querySelectorAll(".pr-table-row").forEach((button) => {
        button.classList.remove("is-active");
      });
      row.classList.add("is-active");
      renderRecentRunSelection(activity).catch((error) => {
        setStatus(`Error: ${error.message}`);
      });
    });
    elements.prTableBody.appendChild(row);
  });

  const firstRow = elements.prTableBody.querySelector(".pr-table-row");
  if (firstRow) {
    firstRow.classList.add("is-active");
    renderRecentRunSelection(activities[0]).catch((error) => {
      setStatus(`Error: ${error.message}`);
    });
  }
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
    const date = formatDate(activity.start_date);
    const isPr = state.prActivityIds.has(activity.id);
    const prLabel = isPr ? "<span class=\"badge\">PR</span>" : "";
    item.innerHTML = `
      <div>
        <div class="activity-title">
          <strong>${activity.name}</strong>
          <a class="strava-link" href="https://www.strava.com/activities/${activity.id}" target="_blank" rel="noopener noreferrer" aria-label="Open activity on Strava">
            <svg viewBox="0 0 24 24" aria-hidden="true">
              <path d="M14 5h5v5" />
              <path d="M10 14L19 5" />
              <path d="M19 14v5h-9a2 2 0 0 1-2-2v-9" />
            </svg>
          </a>
        </div>
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

const PR_TARGETS = [
  { label: "400 m", meters: 400 },
  { label: "800 m", meters: 800 },
  { label: "1 km", meters: 1000 },
  { label: "1 mile", meters: 1609 },
  { label: "5 km", meters: 5000 },
  { label: "10 km", meters: 10000 },
  { label: "Half marathon", meters: 21097 },
  { label: "Marathon", meters: 42195 },
];

function formatPace(seconds, meters) {
  if (!seconds || !meters) return "—";
  const paceSeconds = seconds / (meters / 1000);
  const minutes = Math.floor(paceSeconds / 60);
  const remainder = Math.round(paceSeconds % 60)
    .toString()
    .padStart(2, "0");
  return `${minutes}:${remainder} / km`;
}

function formatTime(seconds) {
  if (!seconds) return "—";
  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  const remainder = Math.round(seconds % 60)
    .toString()
    .padStart(2, "0");
  if (hours > 0) {
    return `${hours}:${minutes.toString().padStart(2, "0")}:${remainder}`;
  }
  return `${minutes}:${remainder}`;
}

function decodePolyline(encoded) {
  if (!encoded) return [];
  let index = 0;
  let lat = 0;
  let lng = 0;
  const coordinates = [];

  while (index < encoded.length) {
    let shift = 0;
    let result = 0;
    let byte = null;
    do {
      byte = encoded.charCodeAt(index++) - 63;
      result |= (byte & 0x1f) << shift;
      shift += 5;
    } while (byte >= 0x20);
    const deltaLat = (result & 1) ? ~(result >> 1) : result >> 1;
    lat += deltaLat;

    shift = 0;
    result = 0;
    do {
      byte = encoded.charCodeAt(index++) - 63;
      result |= (byte & 0x1f) << shift;
      shift += 5;
    } while (byte >= 0x20);
    const deltaLng = (result & 1) ? ~(result >> 1) : result >> 1;
    lng += deltaLng;

    coordinates.push([lat / 1e5, lng / 1e5]);
  }

  return coordinates;
}

function renderPolylineSvg(encodedPolyline) {
  const points = decodePolyline(encodedPolyline);
  if (!points.length) return "<div class=\"map-empty\">NO MAP DATA</div>";

  let minLat = Infinity;
  let maxLat = -Infinity;
  let minLng = Infinity;
  let maxLng = -Infinity;
  points.forEach(([lat, lng]) => {
    minLat = Math.min(minLat, lat);
    maxLat = Math.max(maxLat, lat);
    minLng = Math.min(minLng, lng);
    maxLng = Math.max(maxLng, lng);
  });

  const width = maxLng - minLng || 1;
  const height = maxLat - minLat || 1;

  const svgPoints = points
    .map(([lat, lng]) => {
      const x = (lng - minLng) / width;
      const y = 1 - (lat - minLat) / height;
      return `${x},${y}`;
    })
    .join(" ");

  return `
    <svg class="map-svg" viewBox="0 0 1 1" preserveAspectRatio="xMidYMid meet">
      <polyline points="${svgPoints}" />
    </svg>
  `;
}

function renderMapboxStaticMap(encodedPolyline) {
  if (!encodedPolyline || !config.mapboxToken) {
    return "<div class=\"map-empty\">NO MAP DATA</div>";
  }
  const stroke = "fc5200";
  const lineWidth = 3;
  const lineOpacity = 0.8;
  const overlay = `path-${lineWidth}+${stroke}-${lineOpacity}(${encodedPolyline})`;
  const overlayEncoded = encodeURIComponent(overlay);
  const styleId = "mapbox/light-v11";
  const size = "600x587";
  const padding = 64;
  const url = `https://api.mapbox.com/styles/v1/${styleId}/static/${overlayEncoded}/auto/${size}?padding=${padding}&access_token=${config.mapboxToken}`;
  return `<img class=\"mapbox-image\" src=\"${url}\" alt=\"Route map\" />`;
}

async function renderPrSelection(pr) {
  if (!pr) {
    elements.prMapContent.textContent = "Select a record to view the map.";
    elements.prDetailsContent.textContent = "Select a record to view activity details.";
    return;
  }

  elements.prMapContent.innerHTML = renderMapboxStaticMap(pr.summaryPolyline);
  elements.prDetailsContent.textContent = "Loading activity details...";

  const details = await getActivityDetails(pr.activityId);
  const date = formatDate(details.start_date || pr.startDate);
  const activityDistance = details.distance || pr.distance;
  const activityTime = details.moving_time || pr.elapsedTime;

  elements.prDetailsContent.innerHTML = `
    <div class="pr-details-meta">
      <span class="meta">${details.id}</span>
      <span class="meta">${date}</span>
    </div>
    <div class="pr-details-hero">
      <div class="pr-details-title">
        <a class="activity-link" href="https://www.strava.com/activities/${details.id}" target="_blank" rel="noopener noreferrer">
          ${details.name || pr.activityName || pr.effortName || "Run"}
          <svg viewBox="0 0 24 24" aria-hidden="true">
            <path d="M14 5h5v5" />
            <path d="M10 14L19 5" />
            <path d="M19 14v5h-9a2 2 0 0 1-2-2v-9" />
          </svg>
        </a>
      </div>
    </div>
    <div class="pr-metrics">
      <div class="pr-metric">
        <span class="label">Distance</span>
        <span class="value">${formatDistance(activityDistance)}</span>
      </div>
      <div class="pr-metric">
        <span class="label">Time</span>
        <span class="value">${formatTime(activityTime)}</span>
      </div>
      <div class="pr-metric">
        <span class="label">Avg pace</span>
        <span class="value">${formatPace(activityTime, activityDistance)}</span>
      </div>
      <div class="pr-metric">
        <span class="label">Elevation</span>
        <span class="value">—</span>
      </div>
      <div class="pr-metric">
        <span class="label">Weather</span>
        <span class="value">—</span>
      </div>
      <div class="pr-metric">
        <span class="label">Location</span>
        <span class="value">—</span>
      </div>
    </div>
    <div class="pr-details-footer">
      <div class="pr-footer-title">Personal best</div>
      <div class="pr-footer-grid">
        <div>
          <span class="meta">Distance</span>
          <span>${pr.label || formatDistance(pr.distance)}</span>
        </div>
        <div>
          <span class="meta">Time</span>
          <span>${formatTime(pr.elapsedTime)}</span>
        </div>
        <div>
          <span class="meta">Pace</span>
          <span>${formatPace(pr.elapsedTime, pr.distance)}</span>
        </div>
      </div>
    </div>
  `;
}

function renderPRs(prs) {
  elements.prTableBody.innerHTML = "";

  if (!prs.length) {
    elements.prTableBody.innerHTML = "<div class=\"pr-table-empty\">No PRs found yet.</div>";
    renderPrSelection(null);
    return;
  }

  prs.forEach((pr, index) => {
    const row = document.createElement("button");
    row.type = "button";
    row.className = "pr-table-row";
    row.dataset.index = String(index);
    const effortDate = formatDate(pr.effortDate);
    const distanceLabel = pr.label || formatDistance(pr.distance);
    row.innerHTML = `
      <span>${distanceLabel}</span>
      <span class="meta">${effortDate}</span>
      <span>${formatTime(pr.elapsedTime)}</span>
    `;
    row.addEventListener("click", () => {
      document.querySelectorAll(".pr-table-row").forEach((button) => {
        button.classList.remove("is-active");
      });
      row.classList.add("is-active");
      renderPrSelection(pr).catch((error) => {
        setStatus(`Error: ${error.message}`);
      });
    });
    elements.prTableBody.appendChild(row);
  });

  const firstRow = elements.prTableBody.querySelector(".pr-table-row");
  if (firstRow) {
    firstRow.classList.add("is-active");
    renderPrSelection(prs[0]).catch((error) => {
      setStatus(`Error: ${error.message}`);
    });
  }
}

function matchTarget(distance) {
  const tolerance = 25;
  let closest = null;
  let bestDiff = Infinity;
  PR_TARGETS.forEach((target) => {
    const diff = Math.abs(distance - target.meters);
    if (diff <= tolerance && diff < bestDiff) {
      closest = target;
      bestDiff = diff;
    }
  });
  return closest;
}

async function fetchPRsFromRuns(runs) {
  const cacheMaxAgeMs = CACHE_MAX_AGE_MS;
  const cached = getCachedItem(storageKeys.cachedPrs, storageKeys.cachedPrsAt, cacheMaxAgeMs);
  if (cached && !cached.isStale) {
    return cached.value;
  }

  if (cached?.isStale) {
    setStatus("Refreshing PRs from Strava...");
  }

  const cachedDetails = getCachedDetailsMap();
  const detailsMap = cachedDetails.map;

  const DETAILS_CONCURRENCY = 4;
  const prMap = new Map();
  let processed = 0;

  try {
    const runsNeedingDetails = runs.filter((run) => !detailsMap[run.id]);
    if (runsNeedingDetails.length) {
      let fetched = 0;
      const label = `Fetching activity details... ${fetched}/${runsNeedingDetails.length}`;
      setStatus(label);
      setStatusProgress({
        current: fetched,
        total: runsNeedingDetails.length,
        label,
      });
      await mapWithConcurrency(runsNeedingDetails, DETAILS_CONCURRENCY, async (run) => {
        const details = await fetchWithAuth(`/activities/${run.id}`);
        detailsMap[run.id] = details;
        fetched += 1;
        const updateLabel = `Fetching activity details... ${fetched}/${runsNeedingDetails.length}`;
        setStatus(updateLabel);
        setStatusProgress({
          current: fetched,
          total: runsNeedingDetails.length,
          label: updateLabel,
        });
        return details;
      });
    }

    for (const run of runs) {
      processed += 1;
      const label = `Calculating PRs... ${processed}/${runs.length}`;
      setStatus(label);
      setStatusProgress({
        current: processed,
        total: runs.length,
        label,
      });
      const details = detailsMap[run.id];
      const bestEfforts = details.best_efforts || [];

      bestEfforts.forEach((effort) => {
        const target = matchTarget(effort.distance);
        if (!target) return;
        const existing = prMap.get(target.label);
        if (!existing || effort.elapsed_time < existing.elapsedTime) {
          const effortLabel = effort.name?.trim() || target.label || "Best effort";
          const effortDate = effort.start_date || details.start_date;
          prMap.set(target.label, {
            label: target.label,
            distance: effort.distance,
            elapsedTime: effort.elapsed_time,
            startDate: details.start_date,
            activityId: details.id,
            activityName: details.name,
            effortName: effortLabel,
            effortDate,
            summaryPolyline: details.map?.summary_polyline || "",
          });
        }
      });

    }

    const prs = PR_TARGETS.map((target) => prMap.get(target.label)).filter(Boolean);
    state.detailsMap = detailsMap;
    setCachedDetailsMap(detailsMap);
    setCachedItem(storageKeys.cachedPrs, storageKeys.cachedPrsAt, prs);
    return prs;
  } catch (error) {
    if (cached) {
      setStatus("Using cached PRs (stale). Clear session to refresh.");
      return cached.value;
    }
    throw error;
  }
}

const HEATMAP_COLORS = [
  "#f2f2fb",
  "#dbe8ff",
  "#b7d0ff",
  "#7fb0ff",
  "#4f8cff",
  "#2563eb",
];

function getYearRange(dates) {
  if (!dates.length) return [];
  const years = dates.map((date) => date.getFullYear());
  const minYear = Math.min(...years);
  const currentYear = new Date().getFullYear();
  return Array.from({ length: currentYear - minYear + 1 }, (_, i) => minYear + i);
}

function startOfYear(year) {
  return new Date(Date.UTC(year, 0, 1));
}

function endOfYear(year) {
  return new Date(Date.UTC(year, 11, 31));
}

function formatDayKey(date) {
  return date.toISOString().slice(0, 10);
}

function buildHeatmapData(runs) {
  const totals = new Map();
  runs.forEach((run) => {
    const date = new Date(run.start_date_local);
    const key = formatDayKey(date);
    totals.set(key, (totals.get(key) || 0) + run.distance);
  });
  return totals;
}

function getHeatColor(distance, maxDistance) {
  if (!distance || maxDistance === 0) return HEATMAP_COLORS[0];
  const intensity = distance / maxDistance;
  if (intensity > 0.8) return HEATMAP_COLORS[5];
  if (intensity > 0.6) return HEATMAP_COLORS[4];
  if (intensity > 0.4) return HEATMAP_COLORS[3];
  if (intensity > 0.2) return HEATMAP_COLORS[2];
  return HEATMAP_COLORS[1];
}

function renderHeatmaps(runs) {
  elements.heatmaps.innerHTML = "";

  if (!runs.length) {
    elements.heatmaps.innerHTML = "<p class=\"note\">No runs available for heatmap.</p>";
    return;
  }

  const totals = buildHeatmapData(runs);
  const dates = Array.from(totals.keys()).map((key) => new Date(key));
  const years = getYearRange(dates).reverse();
  const maxDistance = Math.max(...Array.from(totals.values()));
  const today = new Date();
  const todayKey = formatDayKey(new Date(Date.UTC(today.getFullYear(), today.getMonth(), today.getDate())));

  const getTotalWeeks = (year) => {
    const start = startOfYear(year);
    const end = year === today.getFullYear()
      ? new Date(Date.UTC(today.getFullYear(), today.getMonth(), today.getDate()))
      : endOfYear(year);
    const totalDays = Math.round((end - start) / 86400000) + 1;
    const startDay = (start.getUTCDay() + 6) % 7;
    return Math.ceil((startDay + totalDays) / 7);
  };

  const maxWeeks = Math.max(...years.map((year) => getTotalWeeks(year)));

  years.forEach((year) => {
    const yearBlock = document.createElement("div");
    yearBlock.className = "heatmap";

    const header = document.createElement("div");
    header.className = "heatmap-header";

    const yearLabel = document.createElement("span");
    yearLabel.className = "heatmap-year";
    yearLabel.textContent = String(year);

    const yearSummary = document.createElement("span");
    yearSummary.className = "heatmap-summary";

    header.appendChild(yearLabel);
    header.appendChild(yearSummary);

    const body = document.createElement("div");
    body.className = "heatmap-body";

    const yLabels = document.createElement("div");
    yLabels.className = "heatmap-y-labels";
    ["M", "T", "W", "T", "F", "S", "S"].forEach((label) => {
      const span = document.createElement("span");
      span.textContent = label;
      yLabels.appendChild(span);
    });

    const grid = document.createElement("div");
    grid.className = "heatmap-grid";

    const start = startOfYear(year);
    const end = year === today.getFullYear()
      ? new Date(Date.UTC(today.getFullYear(), today.getMonth(), today.getDate()))
      : endOfYear(year);
    const totalDays = Math.round((end - start) / 86400000) + 1;
    const startDay = (start.getUTCDay() + 6) % 7;
    const totalWeeks = Math.ceil((startDay + totalDays) / 7);
    grid.style.gridTemplateColumns = `repeat(${totalWeeks}, 12px)`;
    const gridWidth = totalWeeks * 12 + (totalWeeks - 1) * 4;
    const outerWidth = maxWeeks * 12 + (maxWeeks - 1) * 4;

    const gridOuter = document.createElement("div");
    gridOuter.className = "heatmap-grid-outer";
    gridOuter.style.width = `${outerWidth}px`;

    const gridWrap = document.createElement("div");
    gridWrap.className = "heatmap-grid-wrap";
    gridWrap.style.width = `${gridWidth}px`;

    const xLabels = document.createElement("div");
    xLabels.className = "heatmap-x-labels";
    xLabels.style.gridTemplateColumns = `repeat(${totalWeeks}, 12px)`;

    for (let i = 0; i < totalDays; i += 1) {
      const current = new Date(Date.UTC(year, 0, 1 + i));
      const key = formatDayKey(current);
      const distance = totals.get(key) || 0;
      const weekIndex = Math.floor((startDay + i) / 7);
      const dayIndex = (startDay + i) % 7;
      const cell = document.createElement("div");
      cell.className = "heatmap-cell";
      cell.style.background = getHeatColor(distance, maxDistance);
      cell.style.gridColumn = String(weekIndex + 1);
      cell.style.gridRow = String(dayIndex + 1);
      const tooltipDistance = distance ? formatDistance(distance) : "0 km";
      cell.dataset.tooltip = `${key}: ${tooltipDistance}`;
      cell.dataset.date = key;
      cell.dataset.distance = tooltipDistance;
      grid.appendChild(cell);
    }

    const monthLabels = ["Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"];
    monthLabels.forEach((label, monthIndex) => {
      const monthStart = new Date(Date.UTC(year, monthIndex, 1));
      const dayOffset = Math.round((monthStart - start) / 86400000);
      if (dayOffset < 0 || dayOffset >= totalDays) return;
      const weekIndex = Math.floor((startDay + dayOffset) / 7);
      const span = document.createElement("span");
      span.className = "heatmap-month";
      span.textContent = label;
      span.style.gridColumn = String(weekIndex + 1);
      xLabels.appendChild(span);
    });

    const totalDistance = Array.from({ length: totalDays }, (_, index) => {
      const day = new Date(Date.UTC(year, 0, 1 + index));
      return totals.get(formatDayKey(day)) || 0;
    }).reduce((sum, value) => sum + value, 0);
    const totalKm = totalDistance / 1000;
    const avgKm = totalDays ? totalKm / totalDays : 0;
    yearSummary.textContent = `${totalKm.toFixed(1)}km (${avgKm.toFixed(1)}km/day avg)`;

    const tooltip = document.createElement("div");
    tooltip.className = "heatmap-tooltip";
    tooltip.setAttribute("role", "tooltip");
    tooltip.hidden = true;

    grid.addEventListener("mousemove", (event) => {
      const target = event.target;
      if (!(target instanceof HTMLElement) || !target.classList.contains("heatmap-cell")) {
        tooltip.hidden = true;
        return;
      }
      const rect = target.getBoundingClientRect();
      const containerRect = yearBlock.getBoundingClientRect();
      tooltip.innerHTML = `
        <div>${target.dataset.date}</div>
        <div>${target.dataset.distance}</div>
      `;
      tooltip.style.left = `${rect.left - containerRect.left + rect.width / 2}px`;
      tooltip.style.top = `${rect.top - containerRect.top - 8}px`;
      tooltip.hidden = false;
    });

    grid.addEventListener("mouseleave", () => {
      tooltip.hidden = true;
    });

    gridWrap.appendChild(grid);
    gridWrap.appendChild(xLabels);
    gridOuter.appendChild(gridWrap);
    body.appendChild(yLabels);
    body.appendChild(gridOuter);
    yearBlock.appendChild(header);
    yearBlock.appendChild(body);
    yearBlock.appendChild(tooltip);
    elements.heatmaps.appendChild(yearBlock);
  });
}

async function loadDashboard() {
  setStatus("Loading dashboard from Strava...");
  setStatusProgress({ label: "Loading athlete data...", indeterminate: true });

  const athlete = await fetchWithAuth("/athlete");
  const stats = await fetchWithAuth(`/athletes/${athlete.id}/stats`);
  setStatusProgress({ label: "Loading activities...", indeterminate: true });
  const activities = await fetchAllActivities();

  elements.athleteName.textContent = `${athlete.firstname} ${athlete.lastname}`;

  const runTotals = stats.all_run_totals || stats.recent_run_totals || {};
  elements.totalRuns.textContent = runTotals.count ?? "0";
  elements.totalDistance.textContent = formatDistance(runTotals.distance || 0);
  elements.totalTime.textContent = formatDuration(runTotals.moving_time || 0);

  const runs = activities.filter((activity) => activity.type === "Run");
  const prs = await fetchPRsFromRuns(runs);
  renderPRs(prs);
  state.prs = prs;
  state.recentRuns = runs.slice(0, 10);
  state.longestRuns = [...runs]
    .sort((a, b) => (b.distance || 0) - (a.distance || 0))
    .slice(0, 10);
  state.prActivityIds = new Set(prs.map((pr) => pr.activityId));

  renderStats(runs);
  renderHeatmaps(runs);
  renderAllActivities(activities);

  elements.dashboard.hidden = false;
  setStatus("Dashboard ready.");
  hideStatusProgress();
  setAuthState(true);

  const activeTab =
    document.querySelector(".notable-tabs .tab-btn.is-active")?.dataset.tab || "prs";
  setNotableTab(activeTab);
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
    setAuthState(true);
    window.history.replaceState({}, document.title, redirectUri);
    await loadDashboard();
  } catch (error) {
    setStatus(error.message);
  }
}

function setupEventListeners() {
  elements.connectStrava.addEventListener("click", startAuthFlow);
  elements.logout.addEventListener("click", clearSession);
  if (elements.clearSession) {
    elements.clearSession.addEventListener("click", clearSession);
  }
  document.querySelectorAll(".notable-tabs .tab-btn").forEach((button) => {
    button.addEventListener("click", () => {
      const tab = button.dataset.tab || "prs";
      setNotableTab(tab);
    });
  });
  if (elements.menuToggle && elements.menuPanel) {
    elements.menuToggle.addEventListener("click", () => {
      const isHidden = elements.menuPanel.hidden;
      elements.menuPanel.hidden = !isHidden;
      elements.menuToggle.setAttribute("aria-expanded", String(isHidden));
    });
    document.addEventListener("click", (event) => {
      if (elements.menuPanel.hidden) return;
      const target = event.target;
      if (!(target instanceof Node)) return;
      if (elements.menuPanel.contains(target) || elements.menuToggle.contains(target)) return;
      elements.menuPanel.hidden = true;
      elements.menuToggle.setAttribute("aria-expanded", "false");
    });
  }
}

async function init() {
  await loadConfig();
  setupEventListeners();
  setAuthState(false);

  const token = getToken();
  if (token && !tokenExpired(token)) {
    try {
      setAuthState(true);
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
