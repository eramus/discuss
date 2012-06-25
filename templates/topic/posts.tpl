{{range .}}
	<article style="padding-top: 5px">
		<h1 style="display: none">posted{{if .Username}} by {{.Username}}{{end}} on {{.FTimestamp}}</h1>
		<div style="border-style: dotted; border-width: 1px; padding: 3px">
			<div style="float:left">{{.RTimestamp}}</div>
			<div style="clear: both"></div>
			<span>{{.Post}}</span>
			<footer>
			<div style="float:left; margin-right: 10px;"><a href="/new/post/{{.TId}}/{{.Id}}">Reply</a></div>
			<div style="float:left; margin-right: 10px;">{{.Score}}</div>
			<div style="float:left; margin-right: 10px;"><a href="/bump/post/{{.Id}}">+</a></div>
			<div style="float:left; margin-right: 10px;"><a href="/bury/post/{{.Id}}">-</a></div>
			</footer>
			<div style="clear: both"></div>
		</div>
		{{if .Posts}}
		<div style="padding-left: 30px;">
		{{template "posts.tpl" .Posts}}
		</div>
		{{end}}
	</article>
	<div style="clear: both"></div>
{{end}}