# Chinese Notes Translation Portal

A Go web application and text processing library for translation from Chinese to
English that is adaptable to different dictionaries and corpora, featuring:

1. Chinese-English dictionary word lookup
2. Chinese text segmentation
3. Translation memory (optional)
4. Full text search of a Chinese corpus (optional)
5. Password proection (optional)
6. Integration with a rich JavaScript web client (optional)
7. Integration with a backend SQL databsae (optional)
8. Go module for Chinese text processing

The web app drives the https://chinesenotes.com, https://ntireader.org, and
https://hbreader.org web site and a private translation portal
developed for Fo Guang Shan, working together with the 
[Fo Guang Shan Institute of Humanistic Buddhism](http://fgsihb.org/) and
[Nan Tien Institute](https://www.nantien.edu.au/).

The items above marked as optional require some special set-up, as described
below.

## Quickstart

Install [Go](https://golang.org/doc/install),
clone this repo

```shell
git clone https://github.com/alexamies/chinesenotes-go.git
```

Build the app

```shell
go build
```

Run the web server

```shell
export CNWEB_HOME=.
./chinesenotes-go
```

Navigate to http://localhost:8080

This project contains sufficient data to do minimal integration testing, even if
you have not set up a database or cloned the related dictionary or corpus.

```shell
cd ..
git clone https://github.com/alexamies/chinesenotes.com.git
export CNREADER_HOME=$PWD/chinesenotes.com
```

Then return to this project and re-start the web app:

```shell
cd chinesenotes-go
export CNWEB_HOME=$PWD
./chinesenotes-go
```

Navigate back to http://localhost:8080

The webconfig.yaml file and HTML tempates in /templates allow some additional
customization. The HTML interface is very basic, just enough for minimal
testing. For a real web site you should use HTML templates with JavaScript line
at https://github.com/alexamies/chinesenotes.com

## Integration testing with minimal data

In another terminal

```shell
curl http://localhost:8080/find/?query=邃古
```

You should see JSON encoded data sent back.

## Integration test with real data

To get a fully functioning web app with a JavaScript client and stylized web
pages, generate the HTML files from the corpus by following instructions at

https://github.com/alexamies/chinesenotes.com

Exactly the same process applies to
[NTI Reader](https://github.com/alexamies/buddhist-dictionary)

## Features

### Chinese-English dictionary word lookup

The data/testdict.tsv file gives an example file that illustrates the
structure of the dictionary. The dictionary is indexed by both simplified and
traditional Chinese. Lookup by Chinese word is supported in file mode. 
You will need to install and setup the database to do lookup by English word
Hanyu pinyin. 

### Chinese text segmentation

Given a string of Chinese text, the web app will segment it into words or
multiword expressions contained in the dictionary. This saves a lot of time
for readers who look up many words or discover how the words in a phrase are
grouped, since there are no spaces in Chinese sentences. The presence of the
dictionary files is needed for this. They can be loaded either from the file
system or from the database by the web app.

### Translation memory

Tanslation memory search find the closest matching term based on multiple
criteria, including how many characters match, similarity of the character
order, Pinyin match, and inclusion of the query in the notes. This depends on
compilation of the translation memory index and loading it into the database.

```shell
./cnreader -tmindex
```

See
https://github.com/alexamies/chinesenotes.com/blob/master/bin/cnreader.sh

### Full text search of a Chinese corpus

Full text search of a Chinese corpus allows users to search a monolingual
Chinese corpus. First, you need 

1. Have a corpus following the layout conventions
of Chinese Notes.
2. Compile the index, which computes word and bigram frequencies for each
   document with the cnreader command
3. Load the index files into the database
4. Load the corpus files into Google Cloud Storage

For the corpus structure see
https://github.com/alexamies/chinesenotes.com/tree/master/data/corpus

For compiling the index see
https://github.com/alexamies/chinesenotes.com/blob/master/bin/cnreader.sh

For loading the index into the database see
https://github.com/alexamies/chinesenotes.com/tree/master/index

### Integration with a rich JavaScript web client (optional)

See the web files at 

https://github.com/alexamies/chinesenotes.com/tree/master/web-resources

## FAQ

Q: Why would I use a dictionary and translation memory to translate Chinese
text instead of machine translation?

A: To translate literature, especially classical literature and Buddhist texts,
and to prepare for publishing you will need to thoroughly understand what you
are the source Chinese text.

Q: Can I use the Chinese Notes Translation Portal software for my own project?

A: Yes, please do that. It is also adaptable to your own dictionary, glossary,
and corpus of source text.

## Database Setup

The prepared statements in the Go code assuming a mysql driver. Maria
is compatible with this. The local development instructions assume a Mariadb
database. For full details about the database see the
[Mariadb Documentation](https://mariadb.org/). 

### Mariadb Docker Image

See the documentation at [Mariadb Image 
Documentation](https://hub.docker.com/_/mariadb/) and [Installing and using 
MariaDB via Docker](https://mariadb.com/kb/en/library/installing-and-using-mariadb-via-docker/).

To start a Docker container with Mariadb and connect to it from a MySQL command
line client execute the command below. First, set environment variable 
`MYSQL_ROOT_PASSWORD`. Also, create a directory outside the container to use as a
permanent Docker volume for the database files. In addition, mount volumes for
the tabe separated data to be loaded into Mariadb. See 
[Manage data in Docker](https://docs.docker.com/storage/) and 
[Use volumes](https://docs.docker.com/storage/volumes/) for details on volume
and mount management with Docker.

```shell
MYSQL_ROOT_PASSWORD=[your password]
mkdir mariadb-data
docker run --name mariadb -p 3306:3306 \
  -e MYSQL_ROOT_PASSWORD=$MYSQL_ROOT_PASSWORD -d \
  -v "$(pwd)"/mariadb-data:/var/lib/mysql \
  --mount type=bind,source="$(pwd)"/data,target=/cndata \
  --mount type=bind,source="$(pwd)"/index,target=/cnindex \
  mariadb:10
```

The data in the database is persistent even if the container is deleted. To
restart the database use the command

```shell
docker restart  mariadb
```

To load data from other sources connect to the database container
or start up a mysql-client

```shell
docker exec -it mariadb bash
```

In the container command line

```shell
mysql --local-infile=1 -h localhost -u root -p
```

### Database setup

The first time you run this execute the commands in first_time_setup.sql.

Create the table definitions

```sql
source cndata/chinesenotes.ddl
```

Load sample data

```sql
use cnotest_test;
LOAD DATA LOCAL INFILE 'cndata/grammar.tsv' INTO TABLE grammar CHARACTER SET utf8mb4 LINES TERMINATED BY '\n';
LOAD DATA LOCAL INFILE 'cndata/topics.tsv' INTO TABLE topics CHARACTER SET utf8mb4 LINES TERMINATED BY '\n';
LOAD DATA LOCAL INFILE 'cndata/testdict.tsv' INTO TABLE words CHARACTER SET utf8mb4 LINES TERMINATED BY '\n' IGNORE 1 LINES;
```

The word frequency and translation memory index files are in the index
directory.

Quit from the Maria DB client session

```sql
quit
```

### Run against a databsae

Restart the web application server.


```shell
export DBUSER=app_user
export DBPASSWORD="[your password]"
export DATABASE=cnotest_test
export DBHOST=localhost
./chinesenotes-go
```

From the command line in a new shell you should be able to do a query like

```shell
curl http://localhost:8080/find/?query=antiquity
```

You should see JSON returned.

## Password protecting

To password protect the translation portal set the environment variable
PROTECTED before starting the web server:

```shell
export PROTECTED=true
./chinesenotes-go
```

You will need to add the users manually using SQL statements. There is no
user interface to add users yet.

## Containerize the app and run against a databsae

Build the Docker image for the Go application:

```
sudo docker build -t cn-app-image .
```

Run it locally with minimal features (C-E dictionary lookp only) enabled

```
sudo docker run -it --rm -p 8080:8080 --name cn-app \
  cn-app-image
```

Test basic lookup with curl
```
curl http://localhost:8080/find/?query=你好
```

Set up the database as per the instructions at 
https://github.com/alexamies/chinesenotes.com Then you will be able to run
it locally with all features enabled

```shell
DBUSER=app_user
DBPASSWORD="***"
DATABASE=cnotest_test
docker run -itd --rm -p 8080:8080 --name cn-app --link mariadb \
  -e DBHOST=mariadb \
  -e DBUSER=$DBUSER \
  -e DBPASSWORD=$DBPASSWORD \
  -e DATABASE=$DATABASE \
  -e SENDGRID_API_KEY="$SENDGRID_API_KEY" \
  -e GOOGLE_APPLICATION_CREDENTIALS=/cnotes/credentials.json \
  -e TEXT_BUCKET="$TEXT_BUCKET" \
  --mount type=bind,source="$(pwd)",target=/cnotes \
  cn-app-image
```

Test it

```shell
curl http://localhost:8080/find/?query=hello
```

English queries require a database connection. If everything is working ok, you
should see results like 您好, 哈嘍 and other variations of hello in Chinese.

Debug
```
docker exec -it cn-app bash 
```

Stop it 

```shell
docker stop cn-app
```

Push to Google Container Registry

```shell
docker tag cn-app-image gcr.io/$PROJECT/cn-app-image:$TAG
docker -- push gcr.io/$PROJECT/cn-app-image:$TAG
```

### Go module for Chinese text processing

This GitHub project is a Go module. You can install it with the instructions at
the [Go module reference](https://golang.org/ref/mod), which just involves
importing the APIs and using them. It tries to fail gracefully if you do not
have a database setup and do what it can loading from text files.

Try the example below.

```go
package main

import (
  "context"
  "fmt"
  "github.com/alexamies/chinesenotes-go/dictionary"
  "github.com/alexamies/chinesenotes-go/tokenizer"
)

func main() {
  ctx := context.Background()
  // Works even if you do not have a database
  database, err := dictionary.InitDBCon()
  if err != nil {
    fmt.Printf("unable to connect to database: \n%v\n", err)
  }
  wdict, err := dictionary.LoadDict(ctx, database)
  if err != nil {
    fmt.Printf("unable to load dictionary: \n%v", err)
    return
  }
  tokenizer := tokenizer.DictTokenizer{wdict}
  tokens := tokenizer.Tokenize("戰國時代七雄")
  for _, token := range tokens {
    fmt.Printf("token: %s\n", token.Token)
  }
}
```
