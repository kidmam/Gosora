CREATE TABLE "widgets" (
	`wid` serial not null,
	`position` int not null,
	`side` varchar (100) not null,
	`type` varchar (100) not null,
	`active` boolean DEFAULT 0 not null,
	`location` varchar (100) not null,
	`data` text DEFAULT '' not null,
	primary key(`wid`)
);