/**
 * Database script to DELETE FROMs for corpus index.
 */
USE cnotest_test;

DELETE FROM collection;
DELETE FROM document;
DELETE FROM words;
DELETE FROM topics;
DELETE FROM grammar;
DELETE FROM tmindex_uni_domain;
DELETE FROM tmindex_unigram;
DELETE FROM word_freq_doc;
DELETE FROM bigram_freq_doc;

SHOW WARNINGS;
