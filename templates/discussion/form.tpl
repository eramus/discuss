{{define "content"}}
<form method="post" action="/new/discussion/">
<div id="add_discussion">
	<div style="float: left; width: 100px">title:</div>
	<div style="float: left; width: 150px"><input type="text" name="title" size="20" value="{{.Title}}" /></div>
	<div style="clear: both"></div>
	<div style="float: left; width: 100px">uri:</div>
	<div style="float: left; width: 150px"><input type="text" name="uri" size="20" value="{{.Uri}}" /></div>
	<div style="clear: both"></div>
	<div style="float: left; width: 100px">post:</div>
	<td><textarea name="description">{{.Description}}</textarea></div>
	<div style="clear: both"></div>
	<div style="float: left; width: 100px">keywords:</div>
	<div style="float: left; width: 150px"><input type="text" name="keywords" size="20" value="{{.Keywords}}" /></div>
	<div style="clear: both"></div>
	<div style="float: left; width: 100px"><input type="submit" name="submit" value="Add" /></div>
</div>
</form>
{{end}}