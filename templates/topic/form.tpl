{{define "content"}}
<form method="post" action="/new/topic/{{.Uri}}">
<input type="hidden" name="d_id" value="{{.DId}}" />
<div id="add_topic">
	<div style="float: left; width: 150px">title:</div>
	<div style="float: left; width: 150px"><input type="text" name="title" size="20" value="{{.Title}}" /></div>
	<div style="clear: both"></div>
	<div style="float: left; width: 150px">post:</div>
	<td><textarea name="post">{{.Post}}</textarea></div>
	<div style="clear: both"></div>
	<div style="float: left; width: 150px"><input type="submit" name="submit" value="Add" /></div>
</div>
</form>
{{end}}