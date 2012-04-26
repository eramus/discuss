# discuss

some sort of discussion board

# look at your own risk :)
this is very rough, some of it is a mess and the result of a couple of late night sessions. may cause your machine to explode -- hopefully not though.

lots of work to do. this was mostly just an playground for solr and redis. a discussion board fell out the other end.

# features
can do what a basic forum should do:
	- allow users to login, logout and register + a remember me option
	- view and create topics
	- view and create posts
	- search for things

extra features:
	- some basic +/- voting options for topics and posts
	- users can create discsussions and sub discussions (ie /discuss/cars & /discuss/cars/honda)
	- anon-esque posting
	- threaded discussions
	- its quick (for me atleast)

# prereq
redis -- http://redis.io
noeqd -- https://github.com/bmizerany/noeqd
solr -- http://lucene.apache.org/solr/ (hmm need my config too.. maybe later -- or look at the source and you can figure it out)

