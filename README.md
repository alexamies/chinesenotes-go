# chinesenotes-go

A Go web application and text processing libraries for Chinese-English
dictionary. The web appl drives the Chinese Notes chinesenotes.com and related
web sites. The web app return JSON encoded responses and needs a JavaScript
client to drive it.

## Integration testing with minimal data

This project contains sufficient data to do minimal integration testing, even if
you have not set up a database or cloned the related dictionary or corpus.

To build and run the web app

```shell
go build
./chinesenotes-go
```

In another terminal

```shell
curl http://localhost:8080/find/?query=邃古
```

## Integration test with real data

To do integration testing, clone dictionary and corpus data from 

https://github.com/alexamies/chinesenotes.com

(Exactly the same process applies to
https://github.com/alexamies/buddhist-dictionary and
the private repo for hbreader.org).
Clone that repo and and generate the HTML files from the corpus:

```shell
cd ..
git clone https://github.com/alexamies/chinesenotes.com.git
cd chinesenotes.com
export CNREADER_HOME=$PWD
```

Return to this project and start the web app:

```shell
cd ../chinesenotes-go
export CNWEB_HOME=$PWD
./chinesenotes-go
```

In another terminal

```shell
curl http://localhost:8080/find/?query=邃古
```

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
