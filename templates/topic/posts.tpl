{{range .}}
	<div style="padding-top: 5px">
		<div style="border-style: dotted; border-width: 1px; padding: 3px">
			<div style="float:left">{{.FTimestamp}}</div>
			<div style="clear: both"></div>
			<div style="float:left; margin-right: 10px;"><a href="/bump/post/{{.Id}}">+</a></div>
			<div style="float:left; margin-right: 10px;"><a href="/bury/post/{{.Id}}">-</a></div>
			<div>{{.Post}}</div>
			<div style="float:left; margin-right: 10px;"><a href="/new/post/{{.TId}}/{{.Id}}">Reply</a></div>
			<div style="clear: both"></div>
		</div>
		{{if .Posts}}
		<div style="padding-left: 30px;">
		{{template "posts.tpl" .Posts}}
		</div>
		{{end}}
	</div>
	<div style="clear: both"></div>
{{end}}