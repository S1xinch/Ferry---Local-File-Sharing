package main

import (
	"html/template"
	"net/http"
)

type loginPage struct{ Next, Error string }

func render(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("Referrer-Policy", "no-referrer")
	if err := pages.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var pages = template.Must(template.New("").Parse(`
{{define "head"}}<!doctype html>
<html lang="en"><head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<meta name="robots" content="noindex">
<link rel="icon" href="data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 32 32'%3E%3Crect width='32' height='32' rx='6' fill='%231f5fbf'/%3E%3Cpath d='M11 7.5v17M11 7.5h11M11 16h8' stroke='white' stroke-width='3.5' stroke-linecap='round'/%3E%3C/svg%3E">
<style>
/* Calm Blue: neutral until something needs attention. The accent is blue so a
   button never reads as a status; red (out) and amber (low) are only for
   problems, always with a word and a shape. Every text colour passes WCAG AA. */
:root{--bg:#f6f6f4;--panel:#fff;--line:#e2e1dc;--text:#1d1d1b;--dim:#5f5e59;--accent:#1f5fbf;--accent-text:#fff;
--out:#b42318;--out-bg:#fdecea;--out-fill:#c4320a;--out-ink:#fff;--low:#8a4a00;--low-bg:#fff4e0;--low-fill:#f5b347;--low-ink:#2b1a00;
--ok:#2e8540;--track:#ecebe6;--hover:#f0efea;--sel:#e7eefa;--input:#fff;
--mono:ui-monospace,"Cascadia Mono",Consolas,Menlo,monospace;
--card:var(--panel);--muted:var(--dim);--bad:var(--out)}
@media (prefers-color-scheme:dark){:root{--bg:#161716;--panel:#1f201e;--line:#34352f;--text:#ecebe6;--dim:#a3a29b;--accent:#7aa7ff;--accent-text:#0b1730;
--out:#ff8f80;--out-bg:#3a1c19;--low:#f3b453;--low-bg:#382910;--low-fill:#f3b453;--ok:#4caf68;--track:#2c2d29;--hover:#272824;--sel:#1e2a3e;--input:#191a18;color-scheme:dark}}
*{box-sizing:border-box}
body{margin:0;background:var(--bg);color:var(--text);font:14px/1.45 "Segoe UI",system-ui,-apple-system,sans-serif}
main{max-width:860px;margin:0 auto;padding:24px 16px 48px}
a{color:var(--accent)}
.card{background:var(--panel);border:1px solid var(--line);border-radius:6px}
button,.btn{font:inherit;line-height:1.45;color:var(--text);background:var(--panel);border:1px solid var(--line);border-radius:4px;padding:6px 12px;cursor:pointer;text-decoration:none;display:inline-block;white-space:nowrap}
button:hover,.btn:hover{border-color:var(--dim)}
button.primary,.btn.primary{background:var(--accent);border-color:var(--accent);color:var(--accent-text);font-weight:600}
button.danger{color:var(--out)}
button.small,.btn.small{padding:2px 9px;font-size:13px}
button:disabled{opacity:.45;cursor:default;border-color:var(--line)}
input,select{font:inherit;color:var(--text);background:var(--input);border:1px solid var(--line);border-radius:4px;padding:6px 9px}
:focus-visible,input:focus,select:focus{outline:2px solid var(--accent);outline-offset:1px}
progress,input[type=checkbox]{accent-color:var(--accent)}
.muted{color:var(--dim)}
.error{color:var(--out)}
</style>
{{end}}

{{define "login"}}{{template "head"}}<title>Enter PIN</title></head><body>
<main style="max-width:380px;padding-top:12vh">
<form class="card" method="post" style="padding:24px">
<h1 style="margin:0 0 4px;font-size:22px">Enter the PIN</h1>
<p class="muted" style="margin:0 0 16px">It is shown in the Ferry window on the computer you are connecting to.</p>
<input name="pin" inputmode="numeric" autocomplete="one-time-code" autofocus required maxlength="12"
 style="width:100%;font:600 28px/1 var(--mono);letter-spacing:6px;padding:12px">
{{if .Error}}<p class="error" style="margin:12px 0 0">{{.Error}}</p>{{end}}
<button class="primary" style="width:100%;margin-top:16px;padding:10px">Open</button>
</form>
</main></body></html>{{end}}

{{define "offline"}}{{template "head"}}<title>Website not running</title></head><body>
<main><div class="card" style="padding:24px">
<h1 style="margin:0 0 8px;font-size:22px">The website is not running</h1>
<p class="muted" style="margin:0">Nothing is answering on port {{.Port}} of the host computer. Start the website there, then reload this page.</p>
</div></main></body></html>{{end}}

{{define "list"}}{{template "head"}}<title>{{.Title}}</title>
<style>
nav{display:flex;flex-wrap:wrap;gap:4px;align-items:center;font-size:18px;font-weight:600;margin-bottom:16px}
nav a{text-decoration:none;color:var(--text)}
nav span{color:var(--dim)}
.bar{display:flex;gap:8px;flex-wrap:wrap;margin-bottom:16px;align-items:center}
.bar .grow{flex:1}
table{width:100%;border-collapse:collapse}
td{padding:8px 12px;border-top:1px solid var(--line)}
tr:first-child td{border-top:0}
tbody tr:hover{background:var(--hover)}
td.name a{text-decoration:none;color:var(--text);word-break:break-word}
td.thumb{width:56px;padding-right:0}
td.thumb a{display:block;width:56px;height:56px;text-decoration:none}
td.thumb img{display:block;width:56px;height:56px;object-fit:cover;border-radius:6px;background:var(--bg)}
.tile{display:grid;place-items:center;width:56px;height:56px;border-radius:4px;background:var(--bg);border:1px solid var(--line);color:var(--dim);font:700 11px/1 "Segoe UI",system-ui,sans-serif;letter-spacing:.04em;text-transform:uppercase}
.tile.dir{background:var(--sel);color:var(--accent)}
td.meta{color:var(--dim);white-space:nowrap;font-size:13px;text-align:right;font-variant-numeric:tabular-nums}
td.act{width:1%;white-space:nowrap}
td.act a{font-size:13px}
#drop{border:2px dashed var(--line);border-radius:6px;padding:20px;text-align:center;margin-bottom:16px;background:var(--panel)}
#drop.over{border-color:var(--accent)}
progress{width:100%;margin-top:10px}
.notice{padding:10px 14px;margin-bottom:16px}
@media (max-width:560px){td.meta.date{display:none}}
</style></head><body><main>
<nav>{{range $i, $c := .Crumbs}}{{if $i}}<span>/</span>{{end}}<a href="{{$c.Href}}">{{$c.Name}}</a>{{end}}</nav>
{{if .Notice}}<div class="card notice">{{.Notice}}</div>{{end}}
<div class="bar"><span class="muted grow">{{len .Entries}} items</span>
<a class="btn" href="?zip">Download all as zip</a>
<a class="btn" href="/__ferry/logout">Log out</a></div>
{{if not .ReadOnly}}
<form id="drop" method="post" enctype="multipart/form-data">
<input id="pick" type="file" name="files" multiple hidden>
<p style="margin:0 0 10px">Drop files here to send them to this folder</p>
<button type="button" class="primary" onclick="pick.click()">Choose files</button>
<progress id="prog" value="0" max="100" hidden></progress>
<p id="status" class="muted" style="margin:8px 0 0"></p>
</form>
{{end}}
<div class="card">{{if .Entries}}<table>
{{range .Entries}}<tr>
<td class="thumb"><a href="{{.Href}}" tabindex="-1" aria-hidden="true">{{if .Thumb}}<img src="{{.Href}}" alt="" loading="lazy" decoding="async">{{else if .IsDir}}<span class="tile dir">Folder</span>{{else}}<span class="tile">{{.Badge}}</span>{{end}}</a></td>
<td class="name"><a href="{{.Href}}">{{.Name}}</a></td>
<td class="meta">{{.Size}}</td>
<td class="meta date">{{.Modified}}</td>
<td class="act">{{if .IsDir}}<a href="{{.Href}}?zip">zip</a>{{else}}<a href="{{.Href}}?dl">download</a>{{end}}</td>
</tr>{{end}}
</table>{{else}}<p class="muted" style="padding:14px;margin:0">This folder is empty.</p>{{end}}</div>
{{if not .ReadOnly}}<script>
const drop=document.getElementById('drop'),pick=document.getElementById('pick'),prog=document.getElementById('prog'),status=document.getElementById('status');
function send(files){
  if(!files.length)return;
  const fd=new FormData();for(const f of files)fd.append('files',f);
  const x=new XMLHttpRequest();x.open('POST',location.pathname);
  prog.hidden=false;status.textContent='Uploading '+files.length+' file(s)...';
  x.upload.onprogress=e=>{if(e.lengthComputable)prog.value=e.loaded/e.total*100};
  x.onload=()=>{if(x.status<400)location.href=x.responseURL;else status.textContent='Upload failed: '+x.responseText};
  x.onerror=()=>status.textContent='Upload failed, connection lost.';
  x.send(fd);
}
pick.onchange=()=>send(pick.files);
['dragenter','dragover'].forEach(t=>drop.addEventListener(t,e=>{e.preventDefault();drop.classList.add('over')}));
['dragleave','drop'].forEach(t=>drop.addEventListener(t,e=>{e.preventDefault();drop.classList.remove('over')}));
drop.addEventListener('drop',e=>send(e.dataTransfer.files));
</script>{{end}}
</main></body></html>{{end}}
`))
