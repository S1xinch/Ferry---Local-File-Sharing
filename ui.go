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
<link rel="icon" href="data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 32 32'%3E%3Crect width='32' height='32' rx='6' fill='%231f6f5c'/%3E%3Cpath d='M11 7.5v17M11 7.5h11M11 16h8' stroke='white' stroke-width='3.5' stroke-linecap='round'/%3E%3C/svg%3E">
<style>
:root{--bg:#f6f5f2;--card:#fff;--text:#1d1d1b;--muted:#6b6a65;--line:#e3e1db;--accent:#1f6f5c;--accent-text:#fff;--bad:#a8322a}
@media (prefers-color-scheme:dark){:root{--bg:#161615;--card:#1f1f1d;--text:#ecebe6;--muted:#9a9890;--line:#33322e;--accent:#4fb396;--accent-text:#0d1f1a;--bad:#ff8a80}}
*{box-sizing:border-box}
body{margin:0;background:var(--bg);color:var(--text);font:16px/1.5 system-ui,-apple-system,"Segoe UI",Roboto,sans-serif}
main{max-width:860px;margin:0 auto;padding:24px 16px 48px}
a{color:var(--accent)}
.card{background:var(--card);border:1px solid var(--line);border-radius:8px}
button,.btn{font:inherit;font-weight:600;background:var(--accent);color:var(--accent-text);border:0;border-radius:6px;padding:10px 16px;cursor:pointer;text-decoration:none;display:inline-block}
.btn.plain{background:transparent;color:var(--text);border:1px solid var(--line)}
.muted{color:var(--muted)}
.error{color:var(--bad)}
</style>
{{end}}

{{define "login"}}{{template "head"}}<title>Enter PIN</title></head><body>
<main style="max-width:380px;padding-top:12vh">
<form class="card" method="post" style="padding:24px">
<h1 style="margin:0 0 4px;font-size:22px">Enter the PIN</h1>
<p class="muted" style="margin:0 0 16px">It is shown in the Ferry window on the computer you are connecting to.</p>
<input name="pin" inputmode="numeric" autocomplete="one-time-code" autofocus required maxlength="12"
 style="width:100%;font:600 28px/1 ui-monospace,Menlo,monospace;letter-spacing:6px;padding:12px;border:1px solid var(--line);border-radius:6px;background:var(--bg);color:var(--text)">
{{if .Error}}<p class="error" style="margin:12px 0 0">{{.Error}}</p>{{end}}
<button style="width:100%;margin-top:16px">Open</button>
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
nav span{color:var(--muted)}
.bar{display:flex;gap:8px;flex-wrap:wrap;margin-bottom:16px;align-items:center}
.bar .grow{flex:1}
table{width:100%;border-collapse:collapse}
td{padding:10px 14px;border-top:1px solid var(--line)}
tr:first-child td{border-top:0}
td.name a{text-decoration:none;color:var(--text);word-break:break-word}
td.name .kind{display:inline-block;width:10px;height:10px;border-radius:2px;margin-right:10px;background:var(--line)}
td.name .kind.dir{background:var(--accent)}
td.meta{color:var(--muted);white-space:nowrap;font-size:14px;text-align:right}
td.act{width:1%;white-space:nowrap}
td.act a{font-size:14px}
#drop{border:2px dashed var(--line);border-radius:8px;padding:20px;text-align:center;margin-bottom:16px}
#drop.over{border-color:var(--accent)}
progress{width:100%;margin-top:10px}
.notice{padding:10px 14px;margin-bottom:16px}
@media (max-width:560px){td.meta.date{display:none}}
</style></head><body><main>
<nav>{{range $i, $c := .Crumbs}}{{if $i}}<span>/</span>{{end}}<a href="{{$c.Href}}">{{$c.Name}}</a>{{end}}</nav>
{{if .Notice}}<div class="card notice">{{.Notice}}</div>{{end}}
<div class="bar"><span class="muted grow">{{len .Entries}} items</span>
<a class="btn plain" href="?zip">Download all as zip</a>
<a class="btn plain" href="/__ferry/logout">Log out</a></div>
{{if not .ReadOnly}}
<form id="drop" method="post" enctype="multipart/form-data">
<input id="pick" type="file" name="files" multiple hidden>
<p style="margin:0 0 10px">Drop files here to send them to this folder</p>
<button type="button" onclick="pick.click()">Choose files</button>
<progress id="prog" value="0" max="100" hidden></progress>
<p id="status" class="muted" style="margin:8px 0 0"></p>
</form>
{{end}}
<div class="card">{{if .Entries}}<table>
{{range .Entries}}<tr>
<td class="name"><span class="kind{{if .IsDir}} dir{{end}}"></span><a href="{{.Href}}">{{.Name}}</a></td>
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
