{{define "content"}}
<form method="post" action="/login">
<div id="login">
	<div style="float: left; width: 100px">username:</div>
	<div style="float: left; width: 100px"><input type="text" name="username" size="20" value="{{.Username}}" /></div>
	<div style="clear: both"></div>
	<div style="float: left; width: 100px">password:</div>
	<div style="float: left; width: 100px"><input type="password" name="password" size="20" /></div>
	<div style="clear: both"></div>
	<div style="float: left; margin-right: 5px">remember me?:</div>
	<div style="float: left; width: 100px"><input type="checkbox" name="remember" value="1" {{if .Remember}}checked{{end}} /></div>
	<div style="clear: both"></div>
	<div style="float: left"><input type="submit" name="login" value="Login" /></div>
</div>
</form>
{{end}}