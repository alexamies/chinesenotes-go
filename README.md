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
./chinesenotes-go
```

Navigate to http://localhost:8080

This project contains sufficient data to do minimal integration testing, even if
you have not set up a database or cloned the related dictionary or corpus. To
get a full dictionary and corpus clone the chinesenotes.com repo 

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
DATABASE=cse_dict
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

## Features

### Chinese-English dictionary word lookup

### Chinese text segmentation

### Translation memory


### Full text search of a Chinese corpus

## Password protecting

To password protect the translation portal set the environment variable
PROTECTED before starting the web server:

```shell
export PROTECTED=true
./chinesenotes-go
```

### Integration with a rich JavaScript web client (optional)

### Integration with a backend SQL databsae (optional)


### Go module for Chinese text processing

This GitHub project is a Go module. You can install it with the instructions at
the [Go module reference](https://golang.org/ref/mod), which just involves
importing the APIs and using them, such as the example below.

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

## FAQ

Q: Why would I use a dictionary and translation memory to translate Chinese
text instead of machine translation?

A: To translate literature, especially classical literature and Buddhist texts,
and to prepare for publishing you will need to thoroughly understand what you
are the source Chinese text.

Q: Can I use the Chinese Notes Translation Portal software for my own project?

A: Yes, please do that. It is also adaptable to your own dictionary, glossary,
and corpus of source text.
