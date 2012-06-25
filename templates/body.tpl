<div class="navbar navbar-fixed-top">
<div class="navbar-inner">
<div class="container-fluid">
	<a class="btn btn-navbar" data-toggle="collapse" data-target=".nav-collapse">
		<span class="icon-bar"></span>
		<span class="icon-bar"></span>
		<span class="icon-bar"></span>
	</a>
	<div class="nav-collapse collapse">
		<ul class="nav">
		{{if .Breadcrumbs}}
			<li><a href="/">Home</a></li>
			{{$labels := .Breadcrumbs.Labels}}
			{{range $index, $uri := .Breadcrumbs.Uris}}<li{{if $uri}}{{else}} class="active"{{end}}><a{{if $uri}} href="{{$uri}}"{{else}}{{end}}>{{index $labels $index}}</a></li>
			{{end}}
			<li class="divider-vertical"></li>
			<li>
				<form method="post" action="/search/" class="navbar-search pull-left">
					<input name="search" type="text" class="search-query" placeholder="Search" {{if .Search}} value="{{.Search}}"{{end}}>
				</form>
			</li>
		{{else}}
 			<li class="active"><a>Home</a></li>
		{{end}}
		</ul>
		
		
<ul class="nav pull-right">
{{if .UserData.Id}}
<li class="dropdown">
<a href="#" class="dropdown-toggle" data-toggle="dropdown">hi {{.UserData.Username}}<b class="caret"></b></a>
<ul class="dropdown-menu">
<li><a href="/logout">Logout</a></li>
</ul>
</li>
{{else}}
<li><a href="/register">Register</a></li>
<li><a href="/login">Login</a></li>
{{end}}
</ul>		
		
		
		
		
		
	</div>
</div>
</div>
</div>

<div class="container-fluid">
<div class="row-fluid">

{{if .Subscribed.Labels}}
<div class="span2">
	<ul class="nav nav-list">
		<li class="nav-header">Subscribed</li>
		{{$labels := .Subscribed.Labels}}
		{{range $index, $uri := .Subscribed.Uris}}<li><a href="{{$uri}}">{{index $labels $index}}</a></li>
		{{end}}
	</ul>
</div>
<div class="span10">
{{else}}
<div class="span12">
{{end}}
{{template "content" .ContentData}}
</div>

</div>
</div>