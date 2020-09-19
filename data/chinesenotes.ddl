/*
 * RELATIONAL DATABASE DEFINITIONS FOR Document Search
 * ============================================================================
 */

/*
 * Tables for corpus metadata and index
 *
 * Execute from same directory:
 * > source hbreader.ddl
 */
use cnotest_test;

/*
 * Table listing allowed values for part of speech
 */
CREATE TABLE IF NOT EXISTS grammar (
  english VARCHAR(125) NOT NULL,
  PRIMARY KEY (english)
  )
  CHARACTER SET UTF8
  COLLATE utf8_general_ci
;

/*
 * Table for domain labels
 * id     A unique identifier for the topic
 * word_id    Identifier for the word that the topic relates to
 * simplified:  Simplified Chinese text
 * english:   English text
 * url:     The URL of a page to display information about the topic
 * title:   The title of the page to display information about the topic
 */
CREATE TABLE IF NOT EXISTS topics (
  simplified VARCHAR(125) NOT NULL,
  english VARCHAR(125) NOT NULL,
  url VARCHAR(125),
  title TEXT,
  PRIMARY KEY (simplified, english)
  )
  CHARACTER SET UTF8
  COLLATE utf8_general_ci
;

/*
 * Table for words
 * id     A unique identifier for the word
 * simplified:  Simplified Chinese text for the word
 * traditional: Traditional Chinese text for the word (if different)
 * pinyin:    Hanyu pinyin
 * english:   English text for the word 
 * function:  Grammatical function 
 * concept_cn:  The general concept for the word in Chinese (country, chemical, etc)
 * concept_en:  The general concept for the word in English (country, chemical, etc)
 * topic_cn:  The general topic for the word in Chinese (geography, technology, etc)
 * topic_en:  The general topic for the word in English (geography, technology, etc)
 * parent_cn: The parent for the concept (Chinese)
 * parent_en: The parent for the concept (English)
 * mp3:     Name of an audio file for the word
 * image:   The name of a file for an image illustrating the concept
 * notes:   Encyclopedic notes about the word
 */
CREATE TABLE IF NOT EXISTS words (
  id INT UNSIGNED NOT NULL,
  simplified VARCHAR(255) NOT NULL,
  traditional VARCHAR(255),
  pinyin VARCHAR(255) NOT NULL,
  english VARCHAR(255) NOT NULL,
  grammar VARCHAR(255),
  concept_cn VARCHAR(255),
  concept_en VARCHAR(255),
  topic_cn VARCHAR(125),
  topic_en VARCHAR(125),
  parent_cn VARCHAR(255),
  parent_en VARCHAR(255),
  image VARCHAR(255),
  mp3 VARCHAR(255),
  notes TEXT,
  headword INT UNSIGNED NOT NULL,
  PRIMARY KEY (id),
  FOREIGN KEY (topic_cn, topic_en) REFERENCES topics(simplified, english),
  FOREIGN KEY (grammar) REFERENCES grammar(english),
  INDEX (simplified),
  INDEX (traditional),
  INDEX (english)
  )
  CHARACTER SET UTF8
  COLLATE utf8_general_ci
;

/*
 * Table for illustration licenses
 * name:        The type of license
 * license_full_name: The unabbreviated name of the license
 * license_url:     The URL of the license
 */
CREATE TABLE IF NOT EXISTS licenses (
  name VARCHAR(255) NOT NULL,
  license_full_name VARCHAR(255) NOT NULL,
  license_url VARCHAR(255),
  PRIMARY KEY (name)
  )
  CHARACTER SET UTF8
  COLLATE utf8_general_ci
;

/*
 * Table for illustration authors
 * name:    The name of the creator of the image
 * author_url:  The URL of the home page of the creator of the image
 */
CREATE TABLE IF NOT EXISTS authors (
  name VARCHAR(255),
  author_url VARCHAR(255),
  PRIMARY KEY (name)
  )
  CHARACTER SET UTF8
  COLLATE utf8_general_ci
;

/*
 * Table for illustrations
 * medium_resolution  The file name of a medium resolution image
 * title_zh_cn:     A title in simplified Chinese
 * title_en       A title in English
 * author:        The creator of the illustration
 * license:       The type of license
 * high_resolution:   The file name of a high resolution image
 */
CREATE TABLE IF NOT EXISTS illustrations (
  medium_resolution VARCHAR(255),
  title_zh_cn VARCHAR(255) NOT NULL,
  title_en VARCHAR(255) NOT NULL,
  author VARCHAR(255),
  license VARCHAR(255) NOT NULL,
  high_resolution VARCHAR(255),
  PRIMARY KEY (medium_resolution)/*,
  FOREIGN KEY (author) REFERENCES authors(name),
  FOREIGN KEY (license) REFERENCES licenses(name)*/
  )
  CHARACTER SET UTF8
  COLLATE utf8_general_ci
;

/*
 * Table for collection titles
 */
CREATE TABLE IF NOT EXISTS collection (
  collection_file VARCHAR(256) NOT NULL,
  gloss_file VARCHAR(256) NOT NULL,
	title mediumtext NOT NULL,
	description mediumtext NOT NULL,
	intro_file VARCHAR(256) NOT NULL,
	corpus_name VARCHAR(256) NOT NULL,
	format VARCHAR(256),
  period VARCHAR(256),
  genre VARCHAR(256),
  PRIMARY KEY (`gloss_file`)
	)
    CHARACTER SET UTF8
    COLLATE utf8_general_ci
;

/*
 * Table for document titles
 * plain_text_file - file containing plain text of the document
 * gloss_file - file containing HTML text of the document
 * title - title of the document
 * col_gloss_file - file containing HTML page for the containing collection
 * col_title - title for the containing collection
 * col_plus_doc_title - concatenated title
 */
CREATE TABLE IF NOT EXISTS document (
  plain_text_file VARCHAR(256) NOT NULL,
  gloss_file VARCHAR(256) NOT NULL,
  title mediumtext NOT NULL,
  col_gloss_file VARCHAR(256) NOT NULL,
  col_title mediumtext NOT NULL,
  col_plus_doc_title mediumtext NOT NULL,
  PRIMARY KEY (`gloss_file`)
	)
  CHARACTER SET UTF8
  COLLATE utf8_general_ci
;

/*
 * Table for word frequencies in documents
 * word - Chinese text for the word
 * frequency - the count of words in the document
 * collection - the filename of the HTML Chinese text document
 * document - the filename of the HTML Chinese text document
 * idf - inverse document frequency log[(M + 1) / df(w)]
 */
CREATE TABLE IF NOT EXISTS word_freq_doc (
  word VARCHAR(256) NOT NULL,
  frequency INT UNSIGNED NOT NULL,
  collection VARCHAR(256) NOT NULL,
  document VARCHAR(256) NOT NULL,
  idf FLOAT NOT NULL,
  doc_len INT UNSIGNED NOT NULL,
  PRIMARY KEY (`word`, `document`)
  )
  CHARACTER SET UTF8
  COLLATE utf8_general_ci
;

/*
 * Table for bigram frequencies in documents
 * word - Chinese text for the word
 * frequency - the count of words in the document
 * collection - the filename of the HTML Chinese text document
 * document - the filename of the HTML Chinese text document
 * idf - inverse document frequency log[(M + 1) / df(w)]
 */
CREATE TABLE IF NOT EXISTS bigram_freq_doc (
  bigram VARCHAR(256) NOT NULL,
  frequency INT UNSIGNED NOT NULL,
  collection VARCHAR(256) NOT NULL,
  document VARCHAR(256) NOT NULL,
  idf FLOAT NOT NULL,
  doc_len INT UNSIGNED NOT NULL,
  PRIMARY KEY (`bigram`, `document`)
  )
  CHARACTER SET UTF8
  COLLATE utf8_general_ci
;

/*
 * Table for translation memory index unigrams by character
 * and domain
 * character - Chinese contained in word
 * word - Chinese text for the word
 * domain - The subject domain (concept_en)
 */
CREATE TABLE IF NOT EXISTS tmindex_uni_domain (
  ch VARCHAR(256) NOT NULL,
  word VARCHAR(256) NOT NULL,
  domain VARCHAR(256) NOT NULL,
  PRIMARY KEY (`ch`, `word`, `domain`)
  )
  CHARACTER SET UTF8
  COLLATE utf8_general_ci
;

/*
 * Table for translation memory index unigrams by character
 * character - Chinese contained in word
 * word - Chinese text for the word
 */
CREATE TABLE IF NOT EXISTS tmindex_unigram (
  ch VARCHAR(256) NOT NULL,
  word VARCHAR(256) NOT NULL,
  PRIMARY KEY (`ch`, `word`)
  )
  CHARACTER SET UTF8
  COLLATE utf8_general_ci
;

/*
 * Table for usess for use if the portal is password protected
 */
CREATE TABLE IF NOT EXISTS user (
  UserID INT NOT NULL AUTO_INCREMENT PRIMARY KEY, 
  UserName VARCHAR(100) NOT NULL,
  Email VARCHAR(100), 
  FullName VARCHAR(100) NOT NULL,
  Role VARCHAR(100) NOT NULL DEFAULT "user",
  PasswordNeedsReset TINYINT(1) NOT NULL DEFAULT 1,
  Organization VARCHAR(100),
  Position VARCHAR(100),
  Location VARCHAR(100));

/*
 * Table for hashed passwords if the portal is password protected
 */
CREATE TABLE IF NOT EXISTS passwd (
  UserID INT NOT NULL PRIMARY KEY,
  Password VARCHAR(100),
  CONSTRAINT `fk_user_passwd`
    FOREIGN KEY (UserID) REFERENCES user (UserID)
    ON DELETE CASCADE
    ON UPDATE RESTRICT
  );

/*
 * Table for password reset by email if the portal is password protected
 */
CREATE TABLE IF NOT EXISTS passwdreset (
  Token VARCHAR(100) NOT NULL PRIMARY KEY,
  UserID INT NOT NULL,
  Valid INT NOT NULL DEFAULT 1,
  CONSTRAINT `fk_user_reset`
    FOREIGN KEY (UserID) REFERENCES user (UserID)
    ON UPDATE RESTRICT
  );

/*
 * Table for user sessions if the portal is password protected
 */
CREATE TABLE IF NOT EXISTS session (
  SessionID VARCHAR(100) NOT NULL PRIMARY KEY, 
  UserID INT NOT NULL,
  Active INT NOT NULL DEFAULT 1,
  Authenticated INT NOT NULL DEFAULT 0,
  DailyVisits INT NOT NULL DEFAULT 1,
  Started TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  Updated TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT `fk_user_session`
    FOREIGN KEY (UserID) REFERENCES user (UserID)
    ON DELETE CASCADE
    ON UPDATE RESTRICT
  );
