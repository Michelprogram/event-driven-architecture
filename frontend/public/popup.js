// ---- Kafka WebSocket Log Popup (bottom-left) ----
(function () {
  const MAX_LOGS = 500;

  const LOGS_KEY = "skateshop_kafka_logs";
  function persistLogs(logs) {
    try {
      const serializable = logs.slice(-MAX_LOGS).map((l) => ({
        ...l,
        timestamp:
          l.timestamp instanceof Date ? l.timestamp.toISOString() : l.timestamp,
      }));
      localStorage.setItem(LOGS_KEY, JSON.stringify(serializable));
    } catch {}
  }

  function loadPersistedLogs() {
    try {
      const raw = JSON.parse(localStorage.getItem(LOGS_KEY) || "[]");
      if (!Array.isArray(raw)) return [];
      return raw.map((l) => ({ ...l, timestamp: new Date(l.timestamp) }));
    } catch {
      return [];
    }
  }

  function formatTime(d) {
    return d.toLocaleTimeString("en-US", {
      hour: "2-digit",
      minute: "2-digit",
      second: "2-digit",
      hour12: false,
    });
  }

  function getLevelPill(level) {
    switch (level) {
      case "success":
        return "bg-green-500/20 text-green-300";
      case "warning":
        return "bg-yellow-500/20 text-yellow-300";
      case "error":
        return "bg-red-500/20 text-red-300";
      default:
        return "bg-blue-500/20 text-blue-300";
    }
  }

  function getLevelDot(level) {
    switch (level) {
      case "success":
        return "bg-green-500";
      case "warning":
        return "bg-yellow-500";
      case "error":
        return "bg-red-500";
      default:
        return "bg-blue-500";
    }
  }

  function setupKafkaLogsWidget() {
    // Container
    const wrap = document.createElement("div");
    wrap.className =
      "fixed bottom-4 left-4 z-50 w-[360px] sm:w-[500px] max-h-[500px] bg-gray-900 text-white rounded-lg shadow-2xl overflow-hidden flex flex-col";
    wrap.setAttribute("data-kafka-widget", "true");

    // Header
    const header = document.createElement("div");
    header.className =
      "bg-gray-800 px-4 py-3 flex items-center justify-between border-b border-gray-700";

    const statusGrp = document.createElement("div");
    statusGrp.className = "flex items-center gap-2";

    const statusDot = document.createElement("div");
    statusDot.className = "w-2 h-2 bg-red-400 rounded-full";
    statusDot.setAttribute("aria-label", "connection status");

    const title = document.createElement("span");
    title.className = "font-mono";
    title.textContent = "Kafka: log.central";

    const count = document.createElement("span");
    count.className = "text-xs text-gray-400";
    count.textContent = "(0 events)";

    statusGrp.appendChild(statusDot);
    statusGrp.appendChild(title);
    statusGrp.appendChild(count);

    const btnGrp = document.createElement("div");
    btnGrp.className = "flex items-center gap-2";

    const clearBtn = document.createElement("button");
    clearBtn.className =
      "h-7 w-7 text-gray-400 hover:text-white grid place-items-center rounded hover:bg-white/5";
    clearBtn.title = "Clear";
    clearBtn.setAttribute("aria-label", "Clear");
    clearBtn.textContent = "🗑";

    const toggleBtn = document.createElement("button");
    toggleBtn.className =
      "h-7 w-7 text-gray-400 hover:text-white grid place-items-center rounded hover:bg-white/5";
    toggleBtn.title = "Expand/Collapse";
    toggleBtn.setAttribute("aria-label", "Expand/Collapse");
    toggleBtn.textContent = "▾";

    btnGrp.appendChild(clearBtn);
    btnGrp.appendChild(toggleBtn);

    header.appendChild(statusGrp);
    header.appendChild(btnGrp);

    // Logs container
    const listWrap = document.createElement("div");
    listWrap.className =
      "flex-1 overflow-y-auto p-3 space-y-2 min-h-[200px] max-h-[400px]";

    const emptyState = document.createElement("div");
    emptyState.className = "text-gray-500 text-center py-8 text-sm";
    emptyState.textContent = "No events received yet...";
    listWrap.appendChild(emptyState);

    wrap.appendChild(header);
    wrap.appendChild(listWrap);
    document.body.appendChild(wrap);

    // State
    const state = {
      expanded: true,
      logs: loadPersistedLogs(),
    };

    function render() {
      count.textContent = `(${state.logs.length} events)`;

      // Expand/collapse
      listWrap.style.display = state.expanded ? "" : "none";
      toggleBtn.textContent = state.expanded ? "▾" : "▴";

      // List
      listWrap.innerHTML = "";
      if (state.logs.length === 0) {
        listWrap.appendChild(emptyState);
        return;
      }

      state.logs.forEach((log) => {
        const item = document.createElement("div");
        item.className =
          "bg-gray-800 rounded p-2 border border-gray-700 hover:border-gray-600 transition-colors text-xs font-mono";

        const row = document.createElement("div");
        row.className = "flex items-start gap-2 mb-1";

        const dot = document.createElement("div");
        dot.className = `w-1.5 h-1.5 rounded-full mt-1 ${getLevelDot(
          log.level
        )}`;

        const body = document.createElement("div");
        body.className = "flex-1 min-w-0";

        const metaRow = document.createElement("div");
        metaRow.className = "flex items-center gap-2 mb-1";

        const time = document.createElement("span");
        time.className = "text-gray-500";
        time.textContent = formatTime(log.timestamp);

        const pill = document.createElement("span");
        pill.className = `uppercase px-1.5 py-0.5 rounded text-[10px] ${getLevelPill(
          log.level
        )}`;
        pill.textContent = log.level;

        metaRow.appendChild(time);
        metaRow.appendChild(pill);

        const msg = document.createElement("div");
        msg.className = "text-white mb-1 break-words";
        msg.textContent = log.message;

        body.appendChild(metaRow);
        body.appendChild(msg);

        row.appendChild(dot);
        row.appendChild(body);

        item.appendChild(row);
        listWrap.appendChild(item);
      });
    }

    // Actions
    clearBtn.addEventListener("click", () => {
      state.logs = [];
      persistLogs(state.logs);
      render();
    });

    toggleBtn.addEventListener("click", () => {
      state.expanded = !state.expanded;
      render();
    });

    // WebSocket
    const proto = location.protocol === "https:" ? "wss" : "ws";
    const wsUrl = `${proto}://${location.hostname}:3000/`;
    let ws;

    function connect() {
      try {
        ws = new WebSocket(wsUrl);

        ws.addEventListener("open", () => {
          statusDot.className =
            "w-2 h-2 bg-green-400 rounded-full animate-pulse";
        });

        ws.addEventListener("close", () => {
          statusDot.className = "w-2 h-2 bg-red-400 rounded-full";
          // retry in 2s
          setTimeout(connect, 2000);
        });

        ws.addEventListener("error", () => {
          statusDot.className = "w-2 h-2 bg-red-400 rounded-full";
        });

        ws.addEventListener("message", (e) => {
          let text = "";
          try {
            // logs-service sends plain text; keep robust if JSON comes later
            if (typeof e.data === "string") {
              text = e.data;
            } else {
              text = String(e.data || "");
            }
          } catch {
            text = "[Logs] (unreadable message)";
          }

          const log = {
            id: Date.now() + Math.random(),
            level: "info",
            timestamp: new Date(),
            message: text,
          };

          state.logs.push(log);
          if (state.logs.length > MAX_LOGS) {
            state.logs.splice(0, state.logs.length - MAX_LOGS);
          }
          persistLogs(state.logs);
          render();
        });
      } catch {
        // try again later
        setTimeout(connect, 2000);
      }
    }

    connect();
    render();
  }

  document.addEventListener("DOMContentLoaded", () => {
    // Avoid injecting twice if this file is loaded multiple times
    if (!document.querySelector('[data-kafka-widget="true"]')) {
      setupKafkaLogsWidget();
    }
  });
})();
