# Bibliographical Notes Database Templates

The Bibliographical notes database is an optional part of the Chinese Notes
web app. It is intended to help relate Chinese words and terms to
translations of works in English and other literature discussing Chinese
texts. This directory gives the database schema and sample data.

### Bibliographic Notes Database

Reference Number
This field is the index for the database. Also record the volume, if any, and number of scrolls.

Text Title
Separate fields for Chinese, Hanyu pinyin, and English

Attribution
Authors, translators, or other attribution.

Other Titles
In many cases abbreviated titles of the text are used in reference works and commentaries, which are different to that in the canon. These will be noted. English is used where it is able to be found or, otherwise Hanyu pinyin given.

English translations
There are only some cases where English translations are available.

Summary
There may be notable content in the text, such as a preface or the inclusion of dhāraṇī, that are noted.

Summary
General notes relating to the collection of the data.

Parallels
Where versions in other languages are available.

Commentaries
Historic commentaries on the text. 

References
Modern references to the text will be given in this section.

Chapter titles
Bilingual chapter titles for longer works that are divided into chapters

## Loading into a Database

The format of the data is suitable for any relational database.

### BigQuery

Load the files into BigQuery, following instructions at 
https://cloud.google.com/bigquery/docs/batch-loading-data#bq

From the top level of this project execute the command

```shell
FORMAT=CSV # comma separated variable
PROJECT_ID=[your project] # The alphanumberic ID of your GCP project
DATASET=[your dataset] # To group tables, create a new one if needed
TABLE=bibliographic_notes # The name of the table that will be created
SOURCE=data/bibliographical_notes/taisho_estoteric_section_bibliographic_notes.csv # File to load
SCHEMA=data/bibliographical_notes/bibliographic_notes_schema.json # Schema is in this file
bq load \
--source_format=$FORMAT \
--skip_leading_rows=1 \
${PROJECT_ID}:${DATASET}.${TABLE} \
${SOURCE} \
${SCHEMA}
```

## Querying the Database

Number of records:

```sql
SELECT
  COUNT(reference_no)
FROM
  bibliographic_notes
```

Query for translators, order by number of works translated and also include
the number scrolls:

```sql
SELECT
  attribution_en,
  COUNT(reference_no) AS num_works,
  SUM(no_scrolls) AS num_scrolls
FROM
  bibliographic_notes
GROUP BY
  attribution_en
ORDER BY 
  num_works DESC 
```

Query for dynasty, order by number of works translated and also include
the number scrolls:

```sql
SELECT
  dynasty_en,
  COUNT(reference_no) AS num_works,
  SUM(no_scrolls) AS num_scrolls
FROM
  bibliographic_notes
GROUP BY
  dynasty_en
ORDER BY 
  num_works DESC 
```

Query for dynasty, order by number of works translated and also include
the number scrolls:

```sql
SELECT
  previous_source,
  COUNT(reference_no) AS num_works,
  SUM(no_scrolls) AS num_scrolls
FROM
  bibliographic_notes
GROUP BY
  previous_source
ORDER BY 
  num_works DESC 
```
