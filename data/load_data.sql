/*USE cse_dict;*/
use cnotest_test;

LOAD DATA LOCAL INFILE 'data/grammar.tsv' INTO TABLE grammar CHARACTER SET utf8mb4 LINES TERMINATED BY '\n';
LOAD DATA LOCAL INFILE 'data/topics.tsv' INTO TABLE topics CHARACTER SET utf8mb4 LINES TERMINATED BY '\n';
LOAD DATA LOCAL INFILE 'data/testdict.tsv' INTO TABLE words CHARACTER SET utf8mb4 LINES TERMINATED BY '\n' IGNORE 1 LINES;
SHOW WARNINGS;
