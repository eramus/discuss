{{define "content"}}
<form method="post" action="/new/post/{{.TId}}">
<input type="hidden" name="t_id" value="{{.TId}}" />
{{if .PId}}
<input type="hidden" name="p_id" value="{{.PId}}" />
{{end}}
<div id="add_post">
	<div style="float: left; margin-right: 15px">post:</div>
	<div style="float: left"><textarea name="post">{{.Post}}</textarea></div>
	<div style="clear: both"></div>
	<div style="float: left"><input type="submit" name="add" value="Add" /></div>
	<div style="clear: both"></div>
</div>
</form>
{{end}}