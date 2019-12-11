# chinesenotes-go
Go web application for Chinese-English dictionary.

The source of the dictionary data is at 
https://github.com/alexamies/chinesenotes.com

Clone that repo and copy the words.txt file into the data directory under this
project

```shell
CNOTES_HOME=../chinesenotes.com
mkdir data
cp $CNOTES_HOME/config.yaml .
cp $CNOTES_HOME/data/words.txt data/.
cp $CNOTES_HOME/data/translation_memory_literary.txt data/.
cp $CNOTES_HOME/data/translation_memory_modern.txt data/.
```

That is all that is needed for basic word lookup. A database and corpus can
be added to enable other features.

## Make and Save Go Application Image
The Go app is not needed for chinesenotes.com at the moment but it is use for
other sites (eg. hbreader.org).

Build the Docker image for the Go application:

```
docker build -t cn-app-image .
```

Run it locally with minimal features (C-E dictionary lookp only) enabled
```
docker run -it --rm -p 8080:8080 --name cn-app \
  cn-app-image
```

Test basic lookup with curl
```
curl http://localhost:8080/find/?query=你好
```

Run it locally with all features enabled
```
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

Debug
```
docker exec -it cn-app bash 
```

Push to Google Container Registry

```
docker tag cn-app-image gcr.io/$PROJECT/cn-app-image:$TAG
docker -- push gcr.io/$PROJECT/cn-app-image:$TAG
```
