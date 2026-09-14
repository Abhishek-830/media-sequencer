import { useEffect, useMemo, useState } from "react";
import "./App.css";

const API_BASE = "http://localhost:8080";

function App() {
const [windows, setWindows] = useState([]);
const [loading, setLoading] = useState(true);
const [error, setError] = useState("");

const [selectedWindow, setSelectedWindow] = useState(1);
const [selectedMedia, setSelectedMedia] = useState(null);

const [mediaForm, setMediaForm] = useState({
name: "",
type: "image",
url: "",
duration: 10,
});

const [syncDuration, setSyncDuration] = useState(30);
const [syncMessage, setSyncMessage] = useState("");

const totalCycleSeconds = 5 * 60 * 60;

async function loadWindows() {
try {
const response = await fetch(API_BASE + "/windows");

  if (!response.ok) {
    throw new Error("Failed to load windows");
  }

  const data = await response.json();

  setWindows(data);

  if (
    data.length > 0 &&
    !data.some(function (windowItem) {
      return windowItem.id === selectedWindow;
    })
  ) {
    setSelectedWindow(data[0].id);
  }

  setError("");
} catch (err) {
  setError(
    "Backend se connection nahi ho raha. Check karo Go server port 8080 par chal raha hai."
  );
} finally {
  setLoading(false);
}

}

useEffect(function () {
loadWindows();

const interval = setInterval(function () {
  loadWindows();
}, 2000);

return function () {
  clearInterval(interval);
};

}, []);

const currentWindow = useMemo(
function () {
return windows.find(function (windowItem) {
return windowItem.id === selectedWindow;
});
},
[windows, selectedWindow]
);

async function handleAddMedia(event) {
event.preventDefault();

try {
  const response = await fetch(
    API_BASE + "/windows/" + selectedWindow + "/media",
    {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        name: mediaForm.name,
        type: mediaForm.type,
        url: mediaForm.type === "blank" ? "" : mediaForm.url,
        duration: Number(mediaForm.duration),
      }),
    }
  );

  const data = await response.json();

  if (!response.ok) {
    throw new Error(data.error || "Failed to add media");
  }

  setMediaForm({
    name: "",
    type: "image",
    url: "",
    duration: 10,
  });

  await loadWindows();
} catch (err) {
  setError(err.message);
}

}

async function handleSync() {
if (!selectedMedia) {
setSyncMessage("Pehle koi media select karo.");
return;
}

try {
  const response = await fetch(API_BASE + "/sync", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      media_id: selectedMedia.id,
      duration: Number(syncDuration),
    }),
  });

  const data = await response.json();

  if (!response.ok) {
    throw new Error(data.error || "Sync failed");
  }

  setSyncMessage(
    selectedMedia.name + " sabhi windows par sync ho gaya."
  );
} catch (err) {
  setSyncMessage(err.message);
}

}

function selectMedia(media) {
setSelectedMedia(media);
setSyncMessage("");
}

function handleFormChange(field, value) {
setMediaForm(function (previous) {
return {
...previous,
[field]: value,
};
});
}

return ( <div className="app"> <header className="topbar"> <div> <h1>Media Sequencer</h1> <p>Multi-window synchronized media playback</p> </div>

    <div className="status-badge">
      <span className="status-dot"></span>
      Backend Connected
    </div>
  </header>

  {error && <div className="error-box">{error}</div>}

  <main className="dashboard">
    <section className="section-card">
      <div className="section-heading">
        <div>
          <h2>Display Windows</h2>
          <p>Each window has its own playlist.</p>
        </div>

        <div className="cycle-badge">5 Hour Cycle</div>
      </div>

      {loading ? (
        <div className="loading">Loading windows...</div>
      ) : (
        <div className="windows-grid">
          {windows.map(function (windowItem) {
            return (
              <div
                key={windowItem.id}
                className={
                  "window-card " +
                  (selectedWindow === windowItem.id ? "selected" : "")
                }
                onClick={function () {
                  setSelectedWindow(windowItem.id);
                }}
              >
                <div className="window-header">
                  <h3>{windowItem.name}</h3>
                  <span>{windowItem.playlist.length} items</span>
                </div>

                <div className="playlist-preview">
                  {windowItem.playlist.map(function (media) {
                    return (
                      <button
                        key={media.id}
                        type="button"
                        className={
                          "media-item " +
                          (selectedMedia &&
                          selectedMedia.id === media.id
                            ? "active"
                            : "")
                        }
                        onClick={function (event) {
                          event.stopPropagation();
                          selectMedia(media);
                          setSelectedWindow(windowItem.id);
                        }}
                      >
                        <span className="media-type">
                          {media.type.toUpperCase()}
                        </span>

                        <span className="media-name">
                          {media.name}
                        </span>

                        <span className="media-duration">
                          {media.duration}s
                        </span>
                      </button>
                    );
                  })}
                </div>
              </div>
            );
          })}
        </div>
      )}
    </section>

    <div className="two-column">
      <section className="section-card">
        <div className="section-heading">
          <div>
            <h2>Add Media</h2>
            <p>
              Add media to{" "}
              {currentWindow ? currentWindow.name : "selected window"}.
            </p>
          </div>
        </div>

        <form onSubmit={handleAddMedia} className="media-form">
          <label>
            Window

            <select
              value={selectedWindow}
              onChange={function (event) {
                setSelectedWindow(Number(event.target.value));
              }}
            >
              {windows.map(function (windowItem) {
                return (
                  <option key={windowItem.id} value={windowItem.id}>
                    {windowItem.name}
                  </option>
                );
              })}
            </select>
          </label>

          <label>
            Media Name

            <input
              type="text"
              placeholder="e.g. Product Image"
              value={mediaForm.name}
              onChange={function (event) {
                handleFormChange("name", event.target.value);
              }}
              required
            />
          </label>

          <label>
            Media Type

            <select
              value={mediaForm.type}
              onChange={function (event) {
                handleFormChange("type", event.target.value);
              }}
            >
              <option value="image">Image</option>
              <option value="video">Video</option>
              <option value="blank">Blank</option>
            </select>
          </label>

          {mediaForm.type !== "blank" && (
            <label>
              Media URL

              <input
                type="url"
                placeholder="https://example.com/media.jpg"
                value={mediaForm.url}
                onChange={function (event) {
                  handleFormChange("url", event.target.value);
                }}
                required
              />
            </label>
          )}

          <label>
            Duration (seconds)

            <input
              type="number"
              min="1"
              value={mediaForm.duration}
              onChange={function (event) {
                handleFormChange("duration", event.target.value);
              }}
              required
            />
          </label>

          <button type="submit" className="primary-button">
            + Add Media
          </button>
        </form>
      </section>

      <section className="section-card sync-card">
        <div className="section-heading">
          <div>
            <h2>Sync Playback</h2>
            <p>
              Select any media item and play it across all windows.
            </p>
          </div>
        </div>

        <div className="selected-media">
          <span className="small-label">Selected Media</span>

          {selectedMedia ? (
            <>
              <strong>{selectedMedia.name}</strong>

              <span>
                {selectedMedia.type} · {selectedMedia.duration}s
              </span>
            </>
          ) : (
            <span>No media selected</span>
          )}
        </div>

        <label>
          Sync Duration (seconds)

          <input
            type="number"
            min="1"
            value={syncDuration}
            onChange={function (event) {
              setSyncDuration(event.target.value);
            }}
          />
        </label>

        <button
          type="button"
          className="sync-button"
          onClick={handleSync}
          disabled={!selectedMedia}
        >
          Sync Selected Media
        </button>

        {syncMessage && (
          <div className="sync-message">{syncMessage}</div>
        )}
      </section>
    </div>

    <section className="section-card">
      <div className="section-heading">
        <div>
          <h2>Selected Window Playlist</h2>

          <p>
            {currentWindow
              ? currentWindow.name
              : "No window selected"}{" "}
            · Cycle: {Math.floor(totalCycleSeconds / 3600)} hours
          </p>
        </div>
      </div>

      <div className="playlist-table">
        <div className="table-row table-header">
          <span>Position</span>
          <span>Name</span>
          <span>Type</span>
          <span>Duration</span>
          <span>URL</span>
        </div>

        {currentWindow &&
          currentWindow.playlist.map(function (media, index) {
            return (
              <div className="table-row" key={media.id}>
                <span>#{index + 1}</span>

                <span>{media.name}</span>

                <span>
                  <span className="type-pill">{media.type}</span>
                </span>

                <span>{media.duration}s</span>

                <span className="url-cell">
                  {media.url || "Blank playback"}
                </span>
              </div>
            );
          })}
      </div>
    </section>
  </main>
</div>
);
}

export default App;
