{{define "content"}}
	<a href="/new/post/{{.Id}}">Reply</a>
	<div style="clear: both"></div>
	{{template "posts.tpl" .Posts}}
	<div style="clear: both"></div>
	<a href="/new/post/{{.Id}}">Reply</a>
{{end}}