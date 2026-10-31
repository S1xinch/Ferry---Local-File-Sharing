package main

import "html/template"

func init() {
	template.Must(pages.Parse(panelHTML))
}

const panelHTML = `
{{define "panel-locked"}}{{template "head"}}<title>Ferry</title></head><body>
<main style="max-width:460px;padding-top:12vh"><div class="card" style="padding:24px">
<h1 style="margin:0 0 8px;font-size:22px">Open Ferry again</h1>
<p class="muted" style="margin:0">This page only works when opened by the Ferry app. Open Ferry from your apps and it will bring you back here.</p>
</div></main></body></html>{{end}}

{{define "panel"}}{{template "head"}}<title>Ferry</title>
<style>
main{max-width:760px}
#intro{position:fixed;inset:0;z-index:10;background:var(--bg);display:none;place-items:center;cursor:pointer;transition:opacity .45s ease}
#intro.play{display:grid}
#intro.out{opacity:0}
#intro svg{width:min(440px,88vw);height:auto;overflow:visible}
#intro .pc{fill:none;stroke:var(--text);stroke-width:3;stroke-linejoin:round;stroke-linecap:round}
#intro .screen{fill:var(--card);transition:fill .35s ease}
#intro.arrived .screen.to{fill:var(--accent)}
#intro .file{fill:var(--accent)}
#intro .cargo{transition:opacity .25s ease}
#intro.arrived .cargo{opacity:0}
#intro .landed{fill:var(--accent-text);opacity:0;transform:translateY(-6px);transition:opacity .3s ease .1s,transform .3s ease .1s}
#intro.arrived .landed{opacity:1;transform:none}
#intro .wave{fill:none;stroke:var(--accent);stroke-width:3;stroke-linecap:round;opacity:.55}
#intro .waves{animation:drift 1.2s linear infinite}
#intro .ferry{animation:cross 1.7s cubic-bezier(.45,0,.3,1) .25s both}
#intro .bob{animation:bob .9s ease-in-out infinite alternate;transform-box:fill-box;transform-origin:50% 100%}
#intro .hull{fill:var(--text)}
#intro .cabin{fill:var(--card);stroke:var(--text);stroke-width:2.5}
#intro .stack{fill:var(--accent)}
#intro .name{font:700 22px/1 system-ui,-apple-system,"Segoe UI",sans-serif;fill:var(--text);letter-spacing:.02em;opacity:0;animation:fadein .4s ease 1.6s forwards}
@keyframes cross{from{transform:translateX(0)}to{transform:translateX(158px)}}
@keyframes bob{from{transform:translateY(0) rotate(-2.5deg)}to{transform:translateY(-3px) rotate(2.5deg)}}
@keyframes drift{from{transform:translateX(0)}to{transform:translateX(-40px)}}
@keyframes fadein{to{opacity:1}}
main.enter{animation:enter .45s ease both}
@keyframes enter{from{opacity:0;transform:translateY(8px)}to{opacity:1;transform:none}}
header{display:flex;align-items:center;gap:10px;margin-bottom:20px}
header .logo{width:28px;height:28px;border-radius:6px;background:var(--accent);display:grid;place-items:center;color:var(--accent-text);font-weight:800}
header h1{font-size:20px;margin:0}
header .grow{flex:1}
section.card{padding:20px;margin-bottom:16px}
h2{font-size:17px;margin:0 0 14px}
label.field{display:block;font-weight:600;margin:16px 0 6px}
.row{display:flex;gap:8px}
.row input{flex:1;min-width:0}
input[type=text],input[type=number]{font:inherit;padding:9px 11px;border:1px solid var(--line);border-radius:6px;background:var(--bg);color:var(--text)}
input[type=number]{width:140px}
.seg{display:inline-flex;border:1px solid var(--line);border-radius:6px;overflow:hidden}
.seg button{background:transparent;color:var(--text);border-radius:0;font-weight:600;padding:9px 18px}
.seg button+button{border-left:1px solid var(--line)}
.seg button.on{background:var(--accent);color:var(--accent-text)}
.check{display:flex;gap:10px;align-items:flex-start;margin-top:16px;cursor:pointer}
.check input{margin-top:5px;width:16px;height:16px;accent-color:var(--accent)}
.hint{font-size:14px;color:var(--muted);margin:4px 0 0}
.actions{margin-top:20px;display:flex;gap:10px;align-items:center}
.top{display:flex;gap:12px;align-items:flex-start;justify-content:space-between;flex-wrap:wrap}
.top>div{flex:1;min-width:0}
.what{word-break:break-all;margin:2px 0 0}
.pin{font:700 40px/1.1 ui-monospace,Menlo,Consolas,monospace;letter-spacing:8px;margin:4px 0 0}
.links{display:grid;grid-template-columns:repeat(auto-fit,minmax(260px,1fr));gap:16px;margin-top:20px}
.link{border:1px solid var(--line);border-radius:8px;padding:16px}
.link .url{font:14px ui-monospace,Menlo,Consolas,monospace;word-break:break-all;margin:6px 0 10px}
.link img{display:block;width:180px;height:180px;margin:12px 0 8px;image-rendering:pixelated;background:#fff;padding:8px;border-radius:6px}
.small{font-size:13px;padding:6px 10px}
#log{list-style:none;margin:0;padding:0;font:13px/1.7 ui-monospace,Menlo,Consolas,monospace;max-height:220px;overflow:auto}
#log li{border-top:1px solid var(--line);padding:2px 0}
#log li:first-child{border-top:0}
#closed{display:none;padding:14px 16px;margin-bottom:16px;border-color:var(--bad)}
#update{border-color:var(--accent)}
#update .top{align-items:center}
.label{font-size:13px;font-weight:700;text-transform:uppercase;letter-spacing:.06em;color:var(--muted)}
</style></head><body>
<div id="intro" aria-hidden="true" title="Click to skip">
<svg viewBox="0 0 420 190">
<defs><clipPath id="sea"><rect x="96" y="100" width="228" height="40"/></clipPath></defs>
<g>
<rect class="screen" x="18" y="40" width="78" height="54" rx="5"/>
<rect class="pc" x="18" y="40" width="78" height="54" rx="5"/>
<path class="pc" d="M57 94v14M40 110h34"/>
<path class="file" d="M44 58h12l4 4h14v18H44z"/>
</g>
<g>
<rect class="screen to" x="324" y="40" width="78" height="54" rx="5"/>
<rect class="pc" x="324" y="40" width="78" height="54" rx="5"/>
<path class="pc" d="M363 94v14M346 110h34"/>
<path class="landed" d="M350 58h12l4 4h14v18H350z"/>
</g>
<g clip-path="url(#sea)"><g class="waves">
<path class="wave" d="M80 122q10-7 20 0t20 0 20 0 20 0 20 0 20 0 20 0 20 0 20 0 20 0 20 0 20 0 20 0"/>
<path class="wave" d="M60 134q10-7 20 0t20 0 20 0 20 0 20 0 20 0 20 0 20 0 20 0 20 0 20 0 20 0 20 0 20 0"/>
</g></g>
<g class="ferry"><g class="bob">
<rect class="stack" x="124" y="90" width="7" height="12" rx="1.5"/>
<rect class="cabin" x="112" y="100" width="30" height="13" rx="2"/>
<path class="hull" d="M101 113h52l-7 11h-39z"/>
<path class="file cargo" d="M120 103h5l1.5 1.5h6v6.5h-12.5z"/>
</g></g>
<text class="name" x="210" y="172" text-anchor="middle">Ferry</text>
</svg>
</div>
<main>
<header><div class="logo">F</div><h1>Ferry</h1><span class="muted">{{.Version}}</span><span class="grow"></span>
<button class="btn plain small" id="quit">Quit Ferry</button></header>

<div class="card" id="closed">Ferry has closed. Open the app again to keep sharing.</div>

<section class="card" id="update" hidden>
<div class="top"><div><div class="label">Update ready</div><p class="what" id="update-text"></p></div>
<button type="button" id="update-go">Restart to update</button></div>
</section>

<section class="card" id="setup" hidden>
<h2>What do you want to share?</h2>
<div class="seg"><button type="button" data-mode="files">A folder</button><button type="button" data-mode="site">A website</button><button type="button" data-mode="receive">From another computer</button></div>

<div id="files-opts">
<label class="field" for="folder">Folder</label>
<div class="row"><input type="text" id="folder" placeholder="Choose a folder, or paste its path"><button type="button" class="btn plain" id="browse">Browse</button></div>
<label class="check"><input type="checkbox" id="uploads"><span>Let visitors upload files into this folder</span></label>
<label class="check"><input type="checkbox" id="keeppin"><span>Keep the same PIN when sharing starts again
<span class="hint" style="display:block">Needed if another computer keeps a copy of this folder up to date.</span></span></label>
</div>

<div id="site-opts" hidden>
<label class="field" for="target">Port your website is running on</label>
<input type="number" id="target" min="1" max="65535">
<p class="hint">Start the website on this computer first. Common ports: 3000, 5173, 8000.</p>
</div>

<div id="receive-opts" hidden>
<label class="field" for="server">Address of the other computer</label>
<input type="text" id="server" placeholder="For example 100.64.0.5:8080" autocomplete="off">
<p class="hint">Use the Tailscale address shown on the other computer, with the port Ferry shows there. Both computers need Ferry and Tailscale.</p>
<label class="field" for="rfolder">Folder to keep up to date</label>
<div class="row"><input type="text" id="rfolder" placeholder="Choose a folder, or paste its path"><button type="button" class="btn plain" id="rbrowse">Browse</button></div>
<label class="field" for="rpin">PIN from the other computer</label>
<input type="text" id="rpin" inputmode="numeric" autocomplete="off" maxlength="12" style="width:160px">
<p class="hint">Files in this folder are kept the same as on the other computer, so changes you make to them here are overwritten. Files removed there move to a folder called .ferry-removed here. Files you put in this folder yourself are left alone.</p>
</div>

<label class="check" id="internet-opt"><input type="checkbox" id="internet"><span>Also allow access from outside this Wi-Fi
<span class="hint" style="display:block">Gives a free link that works anywhere. Visitors still need the PIN.</span></span></label>

<p class="error" id="err" hidden></p>
<div class="actions"><button type="button" id="start">Start sharing</button></div>
</section>

<section class="card" id="mirror-live" hidden>
<div class="top"><div><div class="label">Receiving from</div><p class="what" id="mirror-what"></p><p class="hint" id="mirror-status"></p></div>
<button type="button" class="btn plain" id="mirror-stop">Stop receiving</button></div>
</section>

<section class="card" id="live" hidden>
<div class="top"><div><div class="label">Sharing</div><p class="what" id="what"></p></div>
<button type="button" class="btn plain" id="stop">Stop sharing</button></div>
<div style="margin-top:18px"><div class="label">PIN</div><div class="pin" id="pin"></div>
<p class="hint">Visitors type this when they open the link.</p></div>
<div class="links" id="links"></div>
<p class="hint" id="net" hidden></p>
<p class="hint" style="margin-top:16px">Scanning a code logs in without typing the PIN, so only show codes to people you trust. Keep this computer awake while sharing.</p>
</section>

<section class="card"><h2>Activity</h2><ul id="log"></ul></section>
<p class="hint">You can close this tab. Sharing keeps going until you stop it, and opening Ferry again brings this window back.</p>
</main>
<script>
var $ = function (id) { return document.getElementById(id); };
var mode = "files", formFilled = false, lastLinks = "", running = null;

function call(path, body) {
  return fetch(path, {
    method: body === undefined && path === "/api/state" ? "GET" : "POST",
    headers: {"X-Ferry": "1", "Content-Type": "application/json"},
    body: body === undefined ? undefined : JSON.stringify(body)
  }).then(function (r) {
    return r.json().then(function (j) { if (!r.ok) throw new Error(j.error || "Something went wrong"); return j; });
  });
}

function setMode(m) {
  mode = m;
  document.querySelectorAll(".seg button").forEach(function (b) { b.classList.toggle("on", b.dataset.mode === m); });
  $("files-opts").hidden = m !== "files";
  $("site-opts").hidden = m !== "site";
  $("receive-opts").hidden = m !== "receive";
  $("internet-opt").hidden = m === "receive";
  $("start").textContent = m === "receive" ? "Start receiving" : "Start sharing";
}
document.querySelectorAll(".seg button").forEach(function (b) { b.onclick = function () { setMode(b.dataset.mode); }; });

function showError(msg) { msg = msg ? msg.charAt(0).toUpperCase() + msg.slice(1) : ""; $("err").textContent = msg; $("err").hidden = !msg; }

function browseInto(button, input) {
  button.disabled = true;
  call("/api/pick", {}).then(function (r) { if (r.folder) input.value = r.folder; })
    .catch(function (e) { showError(e.message); })
    .finally(function () { button.disabled = false; });
}
$("browse").onclick = function () { browseInto($("browse"), $("folder")); };
$("rbrowse").onclick = function () { browseInto($("rbrowse"), $("rfolder")); };

$("mirror-stop").onclick = function () { call("/api/mirror/stop", {}).then(refresh); };

$("start").onclick = function () {
  showError("");
  $("start").disabled = true;
  call("/api/start", {
    mode: mode, folder: $("folder").value, target: parseInt($("target").value, 10) || 0,
    uploads: $("uploads").checked, internet: $("internet").checked, keepPin: $("keeppin").checked,
    server: $("server").value, mirrorFolder: $("rfolder").value, receivePin: $("rpin").value
  }).then(refresh).catch(function (e) { showError(e.message); })
    .finally(function () { $("start").disabled = false; });
};

$("stop").onclick = function () { call("/api/stop", {}).then(function () { formFilled = false; refresh(); }); };
$("quit").onclick = function () {
  if (running && !confirm("Quitting stops sharing. Quit Ferry?")) return;
  call("/api/quit", {}).finally(function () { $("closed").hidden = false; $("closed").style.display = "block"; });
};

$("update-go").onclick = function () {
  if (running && !confirm("Updating restarts Ferry and stops sharing. Update now?")) return;
  $("update-go").disabled = true;
  call("/api/update", {}).then(function () {
    $("closed").textContent = "Ferry is restarting with the new version. Its window will open in a moment.";
    $("closed").style.display = "block";
  }).catch(function (e) {
    $("update-go").disabled = false;
    $("update-text").textContent = e.message;
  });
};

function el(tag, cls, text) { var e = document.createElement(tag); if (cls) e.className = cls; if (text) e.textContent = text; return e; }

function renderLinks(links) {
  var key = JSON.stringify(links.map(function (l) { return l.url; }));
  if (key === lastLinks) return;
  lastLinks = key;
  var box = $("links");
  box.textContent = "";
  links.forEach(function (l) {
    var card = el("div", "link");
    card.appendChild(el("div", "label", l.label));
    card.appendChild(el("div", "url", l.url));
    var copy = el("button", "btn plain small", "Copy link");
    copy.onclick = function () {
      navigator.clipboard.writeText(l.url).then(function () { copy.textContent = "Copied"; setTimeout(function () { copy.textContent = "Copy link"; }, 1500); });
    };
    card.appendChild(copy);
    var img = el("img"); img.src = l.qr; img.alt = "QR code for " + l.label;
    card.appendChild(img);
    box.appendChild(card);
  });
}

function render(s) {
  running = s.running;
  $("setup").hidden = s.running;
  $("live").hidden = !s.running;
  var mr = s.mirror;
  $("mirror-live").hidden = !mr;
  if (mr) {
    $("mirror-what").textContent = mr.server + " into " + mr.folder;
    $("mirror-status").textContent = mr.lastError ? "Waiting to reconnect: " + mr.lastError
      : mr.lastSync ? "Last checked " + mr.lastSync + ". " + mr.files + " files there, " + mr.changed + " updated and " + mr.removed + " moved to the trash at the last check."
      : "Connecting...";
  }
  var u = s.update || {};
  $("update").hidden = !u.ready;
  if (u.ready) $("update-text").textContent = "Ferry " + u.version + " is downloaded. Restart Ferry to switch to it. Your settings are kept.";
  if (!formFilled) {
    var st = s.settings;
    setMode(st.mode || "files");
    $("folder").value = st.folder || "";
    $("target").value = st.target || 3000;
    $("uploads").checked = !!st.uploads;
    $("internet").checked = !!st.internet;
    $("keeppin").checked = !!st.keepPin;
    $("server").value = st.server || "";
    $("rfolder").value = st.mirrorFolder || "";
    formFilled = true;
  }
  if (s.running) {
    $("what").textContent = s.what;
    $("pin").textContent = s.pin;
    renderLinks(s.links || []);
    $("net").textContent = s.internetStatus || "";
    $("net").hidden = !s.internetStatus;
  } else {
    lastLinks = "";
  }
  var log = $("log");
  log.textContent = "";
  (s.activity.length ? s.activity : ["Nothing yet."]).forEach(function (line) { log.appendChild(el("li", "", line)); });
}

function refresh() {
  return call("/api/state").then(function (s) {
    $("closed").style.display = "none";
    render(s);
  }).catch(function () { $("closed").style.display = "block"; });
}
(function intro() {
  var params = new URLSearchParams(location.search);
  if (!params.has("intro")) return;
  history.replaceState(null, "", "/");
  if (window.matchMedia("(prefers-reduced-motion: reduce)").matches) return;
  var box = $("intro"), main = document.querySelector("main"), done = false;
  function finish() {
    if (done) return;
    done = true;
    box.classList.add("out");
    main.classList.add("enter");
    setTimeout(function () { box.classList.remove("play"); }, 460);
  }
  box.classList.add("play");
  box.onclick = finish;
  setTimeout(function () { box.classList.add("arrived"); }, 1900);
  setTimeout(finish, 2600);
})();
refresh();
setInterval(refresh, 1500);
</script>
</body></html>{{end}}
`
