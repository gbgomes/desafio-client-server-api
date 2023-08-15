Os erros reportados "Erro interno reportado pelo servidor de cotações" ocorrem quando o server não conseguiu acesso ao BD ou à tabela de cotações.

Para corrigir este problema, favor verificar conforme abaixo:


Para que o projeto possa possa ser executado, o server necessita estar devidamante configurado para acesso ao BD, onde a tabela de cotações deve estar criada.

Para isso devem ser executados os seguintes passos:
* Configuração da string de conexão ao BD:
No método gravaCotacoes() a linha 	

    db, err := sql.Open("mysql", "root:#root@tcp(localhost:3306)/goexpert?parseTime=true")

deve ser configurada para o tipo de BD desejado, e string de conexão configurada com usuário e senha, além do nome do database

* criação da tabela de cotações no database especificado na string de conexão ao BD, com o comando abaixo

	CREATE TABLE cotacoes ( <br>
		code char(3) NOT NULL PRIMARY KEY,<br>
		codein char(3),<br>
		name varchar(40),<br>
		high double,<br>
		low double,<br>
		varBid double,<br>
		pctChange double,<br>
		bid double,<br>
		ask double,<br>
		timestamp varchar(10),<br>
		createDate datetime)<br>
