{{define "content"}}
<h1>{{.Title}}</h1>
{{template "posts.tpl" .Posts}}
{{end}}