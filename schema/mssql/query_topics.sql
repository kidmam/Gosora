CREATE TABLE [topics] (
	[tid] int not null IDENTITY,
	[title] nvarchar (100) not null,
	[content] nvarchar (MAX) not null,
	[parsed_content] nvarchar (MAX) not null,
	[createdAt] datetime not null,
	[lastReplyAt] datetime not null,
	[lastReplyBy] int not null,
	[lastReplyID] int DEFAULT 0 not null,
	[createdBy] int not null,
	[is_closed] bit DEFAULT 0 not null,
	[sticky] bit DEFAULT 0 not null,
	[parentID] int DEFAULT 2 not null,
	[ipaddress] nvarchar (200) DEFAULT '0.0.0.0.0' not null,
	[postCount] int DEFAULT 1 not null,
	[likeCount] int DEFAULT 0 not null,
	[attachCount] int DEFAULT 0 not null,
	[words] int DEFAULT 0 not null,
	[views] int DEFAULT 0 not null,
	[css_class] nvarchar (100) DEFAULT '' not null,
	[poll] int DEFAULT 0 not null,
	[data] nvarchar (200) DEFAULT '' not null,
	primary key([tid]),
	fulltext key([content])
);