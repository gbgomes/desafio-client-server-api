/*
CREATE TABLE cotacoes (
	code char(3) NOT NULL PRIMARY KEY,
	codein char(3),
	name varchar(40),
	high double,
	low double,
	varBid double,
	pctChange double,
	bid double,
	ask double,
	timestamp varchar(10),
	createDate datetime)
*/
/*
drop table cotacoes
*/
    
    select * from cotacoes
/*    
    update cotacoes set bid = 9 where code = "USD"
*/

/*
update cotacoes set codein = "kkk", name = "kkk", high = 0, low = 0, varBid = 0, pctChange = 0, bid = 0, ask = 0, timestamp = "kkk" where code = "USD"
*/        







