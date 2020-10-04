#!/bin/bash
## Builds a reverse index for full text search.
## Generates the HTML pages for the web site from corpus text files.
## Run this from the top level directory of the repo.
## DEV_HOME should be set to the location of the Go lang software
## CNREADER_HOME should be set to the location of the staging system
export DEV_HOME=../cnreader
export CNREADER_HOME=`pwd`
export WEB_DIR=translation_portal
export TEMPLATE_HOME=templates
mkdir $WEB_DIR/analysis
mkdir $WEB_DIR/analysis/example_collection
mkdir $WEB_DIR/example_collection
touch index/doc_freq.txt

cd $DEV_HOME
go run cnreader.go
go run cnreader.go -tmindex
cd $CNREADER_HOME
