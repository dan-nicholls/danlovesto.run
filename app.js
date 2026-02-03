const config = {
  clientId: "",
  tokenExchangeUrl: "",
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
  status: document.getElementById("status"),
  dashboard: document.getElementById("dashboard"),
  athleteName: document.getElementById("athleteName"),
  totalRuns: document.getElementById("totalRuns"),
  totalDistance: document.getElementById("totalDistance"),
  totalTime: document.getElementById("totalTime"),
  recentRuns: document.getElementById("recentRuns"),
  prTableBody: document.getElementById("prTableBody"),
  prMapContent: document.getElementById("prMapContent"),
  prDetailsContent: document.getElementById("prDetailsContent"),
  heatmaps: document.getElementById("heatmaps"),
  statsGrid: document.getElementById("statsGrid"),
  allRuns: document.getElementById("allRuns"),
  activityCount: document.getElementById("activityCount"),
};

const state = {
  prActivityIds: new Set(),
};

const STRAVA_API = "https://www.strava.com/api/v3";
const STRAVA_OAUTH = "https://www.strava.com/oauth";

const redirectUri = `${window.location.origin}${window.location.pathname}`;

function setStatus(message) {
  elements.status.textContent = message;
}

function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
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
      waitMs = 1000 * (retryCount + 1);
    }
    setStatus(`Rate limited. Retrying in ${Math.ceil(waitMs / 1000)}s...`);
    await sleep(waitMs);
    return fetchWithAuth(path, { retryCount: retryCount + 1 });
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
  if (cached) {
    if (cached.isStale) {
      setStatus("Using cached activities (stale). Clear session to refresh.");
    }
    return cached.value;
  }

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

  setCachedItem(storageKeys.cachedActivities, storageKeys.cachedActivitiesAt, allActivities);
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

function formatSeconds(seconds) {
  if (!seconds) return "0m";
  const hours = Math.floor(seconds / 3600);
  const minutes = Math.round((seconds % 3600) / 60);
  if (hours > 0) {
    return `${hours}h ${minutes}m`;
  }
  return `${minutes}m`;
}

function renderBarChart(values, labels) {
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
      return `<rect x="${x}" y="${100 - height}" width="${barWidth}" height="${height}" rx="2">
        <title>${label}: ${value.toFixed(1)}</title>
      </rect>`;
    })
    .join("");

  return `
    <svg class="stats-chart" viewBox="0 0 100 100" preserveAspectRatio="none" aria-hidden="true">
      ${bars}
    </svg>
  `;
}

function renderStats(runs) {
  if (!elements.statsGrid) return;
  const totalDistance = runs.reduce((sum, run) => sum + (run.distance || 0), 0);
  const totalTime = runs.reduce((sum, run) => sum + (run.moving_time || 0), 0);
  const totalElev = runs.reduce((sum, run) => sum + (run.total_elevation_gain || 0), 0);
  const longestRun = runs.reduce((max, run) => Math.max(max, run.distance || 0), 0);
  const avgPace = totalDistance > 0 ? totalTime / (totalDistance / 1000) : 0;
  const avgDistance = runs.length ? totalDistance / runs.length : 0;

  const distanceByYear = new Map();
  const runsByWeekday = Array.from({ length: 7 }, () => 0);
  const distanceByMonth = Array.from({ length: 12 }, () => 0);

  runs.forEach((run) => {
    const date = new Date(run.start_date_local);
    const year = date.getFullYear();
    distanceByYear.set(year, (distanceByYear.get(year) || 0) + (run.distance || 0) / 1000);
    const weekday = (date.getDay() + 6) % 7;
    runsByWeekday[weekday] += 1;
    distanceByMonth[date.getMonth()] += (run.distance || 0) / 1000;
  });

  const years = Array.from(distanceByYear.keys()).sort((a, b) => a - b);
  const yearValues = years.map((year) => distanceByYear.get(year));
  const yearLabels = years.map((year) => String(year));

  const weekdayLabels = ["M", "T", "W", "T", "F", "S", "S"];
  const monthLabels = ["Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"];

  const stats = [
    {
      title: "Total distance",
      value: formatDistance(totalDistance),
      subtitle: "All-time runs",
    },
    {
      title: "Total runs",
      value: `${runs.length}`,
      subtitle: "All-time count",
    },
    {
      title: "Total time",
      value: formatSeconds(totalTime),
      subtitle: "Moving time",
    },
    {
      title: "Longest run",
      value: formatDistance(longestRun),
      subtitle: "Single activity",
    },
    {
      title: "Average pace",
      value: avgPace ? `${formatTime(avgPace)} / km` : "—",
      subtitle: "All-time average",
    },
    {
      title: "Avg distance",
      value: formatDistance(avgDistance),
      subtitle: "Per run",
    },
    {
      title: "Distance by year",
      value: `${years.length} yrs`,
      chart: renderBarChart(yearValues, yearLabels),
    },
    {
      title: "Runs by weekday",
      value: "All time",
      chart: renderBarChart(runsByWeekday, weekdayLabels),
    },
    {
      title: "Distance by month",
      value: "All time",
      chart: renderBarChart(distanceByMonth, monthLabels),
    },
  ];

  elements.statsGrid.innerHTML = stats
    .map((stat) => `
      <div class="stats-card">
        <div class="stats-card-header">
          <span>${stat.title}</span>
          <span class="stats-subtitle">${stat.subtitle ?? ""}</span>
        </div>
        <div class="stats-value">${stat.value}</div>
        ${stat.chart || ""}
      </div>
    `)
    .join("");
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
    const isPr = state.prActivityIds.has(activity.id);
    const prLabel = isPr ? "<span class=\"badge\">PR</span>" : "";
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
  if (!points.length) return "<div class=\"map-placeholder\">No map data</div>";

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

function renderPrSelection(pr) {
  if (!pr) {
    elements.prMapContent.textContent = "Select a record to view the map.";
    elements.prDetailsContent.textContent = "Select a record to view activity details.";
    return;
  }

  const date = pr.startDate ? new Date(pr.startDate).toLocaleDateString() : "—";
  elements.prMapContent.innerHTML = renderPolylineSvg(pr.summaryPolyline);
  elements.prDetailsContent.innerHTML = `
    <div class="pr-detail-row"><span class="meta">Activity</span><div>${pr.activityName}</div></div>
    <div class="pr-detail-row"><span class="meta">Activity ID</span><div>#${pr.activityId}</div></div>
    <div class="pr-detail-row"><span class="meta">Date</span><div>${date}</div></div>
    <div class="pr-detail-row"><span class="meta">Distance</span><div>${formatDistance(pr.distance)}</div></div>
    <div class="pr-detail-row"><span class="meta">Time</span><div>${formatTime(pr.elapsedTime)}</div></div>
    <div class="pr-detail-row"><span class="meta">Pace</span><div>${formatPace(pr.elapsedTime, pr.distance)}</div></div>
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
    const effortDate = pr.effortDate ? new Date(pr.effortDate).toLocaleDateString() : "—";
    row.innerHTML = `
      <span>${pr.effortName || pr.label || "Best effort"}</span>
      <span class="meta">${effortDate}</span>
      <span>${formatTime(pr.elapsedTime)}</span>
    `;
    row.addEventListener("click", () => {
      document.querySelectorAll(".pr-table-row").forEach((button) => {
        button.classList.remove("is-active");
      });
      row.classList.add("is-active");
      renderPrSelection(pr);
    });
    elements.prTableBody.appendChild(row);
  });

  const firstRow = elements.prTableBody.querySelector(".pr-table-row");
  if (firstRow) {
    firstRow.classList.add("is-active");
    renderPrSelection(prs[0]);
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
  if (cached) {
    if (cached.isStale) {
      setStatus("Using cached PRs (stale). Clear session to refresh.");
    }
    return cached.value;
  }

  const cachedDetails = getCachedDetailsMap();
  const detailsMap = cachedDetails.map;

  const prMap = new Map();
  let processed = 0;

  for (const run of runs) {
    processed += 1;
    setStatus(`Calculating PRs... ${processed}/${runs.length}`);
    await sleep(250);
    let details = detailsMap[run.id];
    if (!details) {
      details = await fetchWithAuth(`/activities/${run.id}`);
      detailsMap[run.id] = details;
    }
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
  setCachedDetailsMap(detailsMap);
  setCachedItem(storageKeys.cachedPrs, storageKeys.cachedPrsAt, prs);
  return prs;
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

  years.forEach((year) => {
    const yearBlock = document.createElement("div");
    yearBlock.className = "heatmap";

    const header = document.createElement("div");
    header.className = "heatmap-header";
    header.textContent = String(year);

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

    body.appendChild(yLabels);
    body.appendChild(grid);
    yearBlock.appendChild(header);
    yearBlock.appendChild(body);
    yearBlock.appendChild(tooltip);
    elements.heatmaps.appendChild(yearBlock);
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
  const prs = await fetchPRsFromRuns(runs);
  renderPRs(prs);
  state.prActivityIds = new Set(prs.map((pr) => pr.activityId));

  renderRecentRuns(runs.slice(0, 10));
  renderStats(runs);
  renderHeatmaps(runs);
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
