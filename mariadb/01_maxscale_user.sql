CREATE USER 'repl'@'%' IDENTIFIED BY 'repl';
GRANT ALL ON *.* TO 'repl'@'%';

CREATE USER 'cdcuser'@'%' IDENTIFIED BY 'cdc';
GRANT ALL ON *.* TO 'cdcuser'@'%';

RESET MASTER;
