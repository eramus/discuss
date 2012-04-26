{{define "content"}}
<div><b>found:</b>{{.NumFound}}</div>
{{if .Discussions}}
<h2>Discussions</h2>
{{range .Discussions}}
<div style="margin: 5px 0px 5px 0px; padding: 5px 0px 5px 0px">
	<div style="border-style: dotted; border-width: 1px; padding: 3px">
		<div style="float:left; margin-right: 10px;"><b><a href="/discuss{{range .Uri}}/{{.}}{{end}}">{{.Title}}</a></b></div>
		<div><b>{{range .Uri}}/{{.}}{{end}}</b></div>
		<div style="clear: both"></div>
		<div style="float:left">{{.Description}}</div>
		<div style="clear: both"></div>
		<div style="float:left; margin-right: 15px;"><b>topics:</b>&nbsp;{{.Topics}}</div>
		<div style="float:left"><b>subscribed:</b>&nbsp;{{.Subscribed}}</div>
		<div style="clear: both"></div>
	</div>
</div>
{{end}}
<div style="clear: both"></div>
{{end}}

{{if .Posts}}
<h2>Posts</h2>
{{range .Posts}}
<div style="margin: 5px 0px 5px 0px; padding: 5px 0px 5px 0px">
	<div style="border-style: dotted; border-width: 1px; padding: 3px">
		<div style="float:left; margin-right: 10px;"><b><a href="/topic/{{.TId}}#{{.Id}}">{{if .Title}}{{.Title}}{{else}}view post{{end}}</a></b></div>
		<div style="clear: both"></div>
		<div style="float:left">{{.Post}}</div>
		<div style="clear: both"></div>
	</div>
</div>
{{end}}
<div style="clear: both"></div>
{{end}}


<h2>Search Again</h2>
<form method="post" action="/search/">
<div id="search"><input type="text" name="search" size="20" value="{{.Query}}" />&nbsp;<input type="submit" name="submit" value="Find" /></div>
</form>
<div style="clear: both"></div>
<h2>Add Discussion</h2>
<form method="post" action="/new/discussion/">
<div id="add_discussion">
	<div style="float: left; width: 100px">title:</div>
	<div style="float: left; width: 150px"><input type="text" name="title" size="20" value="{{.Query}}" /></div>
	<div style="clear: both"></div>
	<div style="float: left; width: 100px">uri:</div>
	<div style="float: left; width: 150px"><input type="text" name="uri" size="20" value="" /></div>
	<div style="clear: both"></div>
	<div style="float: left; width: 100px">post:</div>
	<td><textarea name="description"></textarea></div>
	<div style="clear: both"></div>
	<div style="float: left; width: 100px">keywords:</div>
	<div style="float: left; width: 150px"><input type="text" name="keywords" size="20" value="" /></div>
	<div style="clear: both"></div>
	<div style="float: left; width: 100px"><input type="submit" name="submit" value="Add" /></div>
</div>
</form>
{{end}}