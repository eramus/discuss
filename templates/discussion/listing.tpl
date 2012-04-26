{{define "content"}}
<a href="/new/topic/{{.Uri}}">New Topic</a>
{{range .Topics}}
<div style="margin: 5px 0px 5px 0px; border-style: dotted; border-width: 1px; padding: 3px">
	<div style="float:left; margin-right: 10px;"><a href="/bump/topic/{{.Id}}">+</a></div>
	<div style="float:left; margin-right: 10px;"><a href="/bury/topic/{{.Id}}">-</a></div>
	<div><a href="/topic/{{.Id}}">{{.Title}}</a></div>
	<span>posts: {{.NumPosts}}</span>
	<span>last post: {{.LastPost}}</span>
</div>
{{end}}
<a href="/new/topic/{{.Uri}}">New Topic</a>
{{end}}