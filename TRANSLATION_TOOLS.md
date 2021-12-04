# translatetools
A prototype for translation of Chinese docs with machine translation APIs and
related processing of output text for style.

## DeepL

Sign up for a DeepL account and get an API key. Set the key as an environment
variable:

```shell
export DEEPL_AUTH_KEY="your key"
```

## Google Translation API

### API Setup
Create Google Cloud Platform project with billing, enable the Translate API, and
create a service account and key, as explained in the
[Translation API Setup](https://cloud.google.com/translate/docs/setup).

Set an environment variable with the location of the key file:

```shell
export GOOGLE_APPLICATION_CREDENTIALS=service-account-file.json
```

Install the Go client API:
```shell
go get -u cloud.google.com/go/translate
```

Install and initialize the Cloud SDK.

### Glossary

Follow instructions at
https://cloud.google.com/translate/docs/advanced/glossary

Upload the glossary to GCS using the command:

```shell
GLOSSARY_BUCKET=[your bucket name]
GLOSSARY_FILE=data/glossary/[your CSV glossary]
gsutil cp $GLOSSARY_FILE gs://${GLOSSARY_BUCKET}/
```

Check contents of the bucket:

```shell
gsutil ls -r gs://${GLOSSARY_BUCKET}/**
```

Edit the file `glossary_request.json`. Create the glossary sources with the
command

```shell
PROJECT_ID=[your project]
curl -X POST \
  -H "Authorization: Bearer "$(gcloud auth application-default print-access-token) \
  -H "Content-Type: application/json; charset=utf-8" \
  -d @glossary_request.json \
  "https://translation.googleapis.com/v3/projects/${PROJECT_ID}/locations/us-central1/glossaries"
```

check status of operation

```shell
OPERATION_ID="20211204-11321638646361-6195e5b8-0000-215f-920a-582429aa5674"
curl -X GET \
-H "Authorization: Bearer "$(gcloud auth application-default print-access-token) \
"https://translation.googleapis.com/v3/projects/${PROJECT_ID}/locations/us-central1/operations/${OPERATION_ID}"
```

Check glossary has been created successfully by listing glossaries:

```shell
curl -X GET \
-H "Authorization: Bearer "$(gcloud auth application-default print-access-token) \
"https://translation.googleapis.com/v3/projects/${PROJECT_ID}/locations/us-central1/glossaries"
```

Delete a glossary

```shell
curl -X DELETE \
-H "Authorization: Bearer "$(gcloud auth application-default print-access-token) \
"https://translation.googleapis.com/v3/projects/${PROJECT_ID}/locations/us-central1/glossaries/test-fgdb-glossary"
```
