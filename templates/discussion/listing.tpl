{{define "content"}}
<a href="/new/topic/{{.Uri}}">New Topic</a>
{{range .Topics}}
<div id="topic" style="margin: 5px 0px 5px 0px; border-style: dotted; border-width: 1px; padding: 3px">
	<article>
		<div style="float: left">
			<div style="width: 15px; padding: 2px 2px 2px 2px; font-size: large; margin-right: 5px;"><a href="/bump/topic/{{.Id}}"><img src="/images/plus.png" /></a></div>
			<div style="width: 15px; padding: 2px 2px 2px 2px; font-size: large; margin-right: 5px;"><a href="/bury/topic/{{.Id}}"><img src="/images/minus.png" /></a></div>
		</div>
		<div style="float: left">
			<header>
				<div style="float: left; font-size: large; margin-right:5px">[{{.Score}}]</div>
				<h2 style="float: left; font-size: large; margin: 0px 0px 0px 0px"><a href="/topic/{{.Id}}">{{.Title}}</a></h2>
			<header>
			<div style="clear: both"></div>
			<footer>
				<b>posts:</b> {{.NumPosts}}&nbsp;<b>last post:</b> {{.LastPost}}&nbsp;<b>users:</b> {{.Users}}
			</footer>
		</div>
	</article>
	<div style="clear: both"></div>
</div>
{{end}}
<a href="/new/topic/{{.Uri}}">New Topic</a>
{{end}}