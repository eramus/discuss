{{define "content"}}
<form method="post" action="/register">
<div id="register">
	<div style="float: left; width: 150px">username:</div>
	<div style="float: left; width: 150px"><input type="text" name="username" size="20" value="{{.Username}}" /></div>
	<div style="clear: both"></div>
	<div style="float: left; width: 150px">password:</div>
	<div style="float: left; width: 150px"><input type="password" name="password" size="20" /></div>
	<div style="clear: both"></div>
	<div style="float: left; width: 150px">confirm:</div>
	<div style="float: left; width: 150px"><input type="password" name="confirm_password" size="20" /></div>
	<div style="clear: both"></div>
	<div style="float: left; width: 150px">email:</div>
	<div style="float: left; width: 150px"><input type="text" name="email" size="20" value="{{.Email}}" /></div>
	<div style="clear: both"></div>
	<div style="float: left; width: 150px" colspan="2"><input type="submit" name="login" value="Login" /></div>
</div>
</form>
{{end}}