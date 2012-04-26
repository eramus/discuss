<div id="user_data" style="float:right">
{{if .UserData.Id}}hi {{.UserData.Username}}&nbsp;<a href="/logout">Logout</a>{{else}}<a href="/register">Register</a>&nbsp;<a href="/login">Login</a>{{end}}
</div>
<div id="breadcrumbs" style="float:left">
{{if .Breadcrumbs}}
<a href="/">Home</a>
{{$labels := .Breadcrumbs.Labels}}
{{range $index, $uri := .Breadcrumbs.Uris}} &raquo; {{if $uri}}<a href="{{$uri}}">{{index $labels $index}}</a>{{else}}{{index $labels $index}}{{end}}{{end}}
{{else}}
Home
{{end}}
</div>
<div style="clear: both"></div>
{{if .Subscribed}}
<div style="float:left; width: 10%; margin-top: 20px">
<span><b>Subscribed</b></span>
<div style="clear: both"></div>
{{$labels := .Subscribed.Labels}}
{{range $index, $uri := .Subscribed.Uris}} &raquo; <a href="{{$uri}}">{{index $labels $index}}</a><br />{{end}}
</div>
{{end}}
<div id="content" style="float:left; width: 88%; margin-top: 20px">
{{template "content" .ContentData}}
</div>