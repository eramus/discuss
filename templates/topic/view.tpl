{{define "content"}}
<span style="margin: 2px"><a href="/new/post/{{.Id}}">Reply</a></span>
<div style="clear: both"></div>
{{template "posts.tpl" .Posts}}
<div style="clear: both"></div>
<span style="margin-top: 10px"><a href="/new/post/{{.Id}}">Reply</a></span>
{{end}}